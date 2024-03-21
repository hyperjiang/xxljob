// Package xxljob implements a golang executor for xxl-job.
// It provides functions to execute jobs, get job logs, and report job execution status.
// An executor can run multiple jobs concurrently, but each job is run in serial mode.
package xxljob

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	resty "github.com/go-resty/resty/v2"
	"github.com/hyperjiang/scheduler"
)

const (
	logPrefix         = "[xxl-job] "
	accessTokenHeader = "XXL-JOB-ACCESS-TOKEN"

	successCode = 200
	failureCode = 500
)

// Executor is responsible for executing jobs.
type Executor struct {
	Options

	registry *RegistryParam
	cli      *resty.Client
	srv      *http.Server
	// key is handler name, value is handler func
	handlers sync.Map
	// current running jobs. key is job id, value is job instance.
	// if a job is finished, it should be removed from this map.
	jobs         sync.Map
	callbackChan chan CallbackParam
	notifier     *scheduler.Scheduler
	registrar    *scheduler.Scheduler
}

// NewExecutor creates a new executor.
func NewExecutor(opts ...Option) *Executor {
	e := &Executor{
		Options: NewOptions(opts...),
	}

	e.registry = &RegistryParam{
		RegistryGroup: "EXECUTOR",
		RegistryKey:   e.AppName,
		RegistryValue: fmt.Sprintf("http://%s:%d", LocalIP(), e.Port),
	}

	// Init http client.
	e.cli = resty.New().
		SetBaseURL(e.Host).
		SetTimeout(e.ClientTimeout).
		SetHeader("Content-Type", "application/json")
	if e.AccessToken != "" {
		e.cli.SetHeader(accessTokenHeader, e.AccessToken)
	}

	// Init http server.
	e.setupRoutes()
	e.srv = &http.Server{
		Addr:         fmt.Sprintf(":%d", e.Port),
		IdleTimeout:  e.IdleTimeout,
		ReadTimeout:  e.ReadTimeout,
		WriteTimeout: e.WriteTimeout,
	}

	// Send result notifications to xxl-job server periodically.
	e.callbackChan = make(chan CallbackParam, e.CallbackBufferSize)
	e.notifier = scheduler.New("xxljob_callback", e.notifyResult, e.CallbackInterval)
	e.notifier.Start()

	// xxl-job server does not check executor's health, so we need to register periodically in order to keep alive.
	e.registrar = scheduler.New("xxljob_register", e.register, e.RegisterInterval)
	e.registrar.Start()

	return e
}

// post conduct a post request and parse the response.
func (e *Executor) post(endpoint string, data interface{}, res interface{}) error {
	resp, err := e.cli.R().SetBody(data).SetResult(res).Post(endpoint)

	body := resp.String()
	size := resp.Size()
	if size > e.SizeLimit {
		body = "omitted"
	}

	e.Logger.Info(logPrefix+"[%d][%s][%s] url: %s, res: %s",
		resp.StatusCode(),
		ReadableSize(size),
		TruncateDuration(resp.Time()),
		resp.Request.URL,
		body,
	)

	return err
}

// register registers the executor.
func (e *Executor) register() error {
	var res Response

	err := e.post("/api/registry", e.registry, &res)
	if err != nil {
		e.Logger.Error(logPrefix+"register executor failed: %v", err)
	}

	return err
}

// deregister deregisters the executor.
func (e *Executor) deregister() error {
	var res Response
	err := e.post("/api/registryRemove", e.registry, &res)
	if err != nil {
		e.Logger.Error(logPrefix+"deregister executor failed: %v", err)
	}

	return err
}

// callback reports job execution status to xxl-job server.
func (e *Executor) callback(callbacks []CallbackParam) error {
	var res Response
	err := e.post("/api/callback", callbacks, &res)
	if err != nil {
		e.Logger.Error(logPrefix+"callback failed: %v", err)
	}

	return err
}

// notifyResult sends job execution results to xxl-job server asynchronously.
func (e *Executor) notifyResult() error {
	var callbacks []CallbackParam

	if len(e.callbackChan) == 0 {
		return nil
	}

	for {
		select {
		case cb := <-e.callbackChan:
			callbacks = append(callbacks, cb)
		default:
			// No more data to receive.
			return e.callback(callbacks)
		}
	}
}

// Start starts the executor and register itself to the xxl-job server.
func (e *Executor) Start() error {
	if err := e.register(); err != nil {
		return err
	}

	// Run our server in a goroutine so that it doesn't block
	errChan := make(chan error, 1)
	go func() {
		e.Logger.Info("http server listen and serve on :%d", e.Port)
		errChan <- e.srv.ListenAndServe()
	}()

	// Intercept interrupt signals.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, e.interruptSignals...)

	// Wait for error or shutdown signal.
	select {
	// If http server fail to start.
	case err := <-errChan:
		e.Logger.Error("fail to start http server: %s", err.Error())
		return err
	// If we receive an interrupt signal, gracefully shutdown server.
	case sig := <-sigChan:
		e.Logger.Info("interrupt signal received: %v", sig)

		// Create a deadline to wait for graceful shutdown.
		ctx, cancel := context.WithTimeout(context.Background(), e.WaitTimeout)
		defer cancel()

		return e.srv.Shutdown(ctx)
	}
}

// Stop stops the executor.
func (e *Executor) Stop() error {
	e.registrar.Stop()
	_ = e.deregister()

	e.jobs.Range(func(k interface{}, v interface{}) bool {
		job := v.(*Job)
		e.stopJob(job)
		return true
	})

	close(e.callbackChan)
	e.notifier.Stop()

	return nil
}

// GetJobHandler retrieves the job handler for a given name.
func (e *Executor) GetJobHandler(name string) JobHandler {
	v, ok := e.handlers.Load(name)
	if !ok {
		return nil
	}

	return v.(JobHandler)
}

// AddJobHandler registers a job handler by name.
func (e *Executor) AddJobHandler(name string, h JobHandler) {
	e.handlers.Store(name, h)
}

// RemoveJobHandler removes a job handler by name.
func (e *Executor) RemoveJobHandler(name string) {
	e.handlers.Delete(name)
}

// TriggerJob triggers a job.
// It will return error if handler does not exist or log id is duplicate.
func (e *Executor) TriggerJob(params RunParam) error {
	// Check if there is a job with same id running.
	job := e.getJob(params.JobID)
	if job != nil && job.LogID == params.LogID {
		return errors.New("duplicate log id")
	}

	switch params.ExecutorBlockStrategy {
	case DiscardLater:
		// If there is a job with same id running, this request will be discarded and marked as failed.
		if job != nil {
			e.Logger.Info(logPrefix+"[%d:%d] is still running", job.ID, job.LogID)
			return errors.New("a job of same id is already running")
		}
	case CoverEarly:
		// If there is a job with same id running, we will terminate it,
		// and then creates and runs a new one.
		if job != nil {
			e.stopJob(job)
		}
		fallthrough
	case SerialExecution:
		// Default mode, put the job into the FIFO queue and runs in serial mode.
		fallthrough
	default:
	}

	newJob, err := e.newJob(params)
	if err != nil {
		return err
	}
	go e.runJob(newJob)

	return nil
}

// getJob tries to get the current running job by id.
func (e *Executor) getJob(id int) *Job {
	v, ok := e.jobs.Load(id)
	if !ok {
		return nil
	}

	return v.(*Job)
}

// addJob saves the current running job in the job map.
func (e *Executor) addJob(job *Job) {
	e.jobs.Store(job.ID, job)
}

// stopJob stops and removes a job.
func (e *Executor) stopJob(job *Job) {
	e.Logger.Info(logPrefix+"[%d:%d] job is stopped and removed", job.ID, job.LogID)
	job.Stop()
	e.jobs.Delete(job.ID)
}

// newJob creates a new job instance and starts watching its execution.
func (e *Executor) newJob(params RunParam) (*Job, error) {
	handler := e.GetJobHandler(params.ExecutorHandler)
	if handler == nil {
		return nil, errors.New("job handler not found")
	}

	param := JobParam{
		Params:        params.ExecutorParams,
		ShardingIndex: params.BroadcastIndex,
		ShardingTotal: params.BroadcastTotal,
	}

	job := &Job{
		ID:      params.JobID,
		LogID:   params.LogID,
		Name:    params.ExecutorHandler,
		Handle:  handler,
		Param:   param,
		Timeout: params.ExecutorTimeout,
		done:    make(chan error, 1),
	}

	go e.watch(job)

	return job, nil
}

// watch waits for the job execution result and push it to the callback queue.
func (e *Executor) watch(job *Job) {
	err := <-job.done

	job.EndTime = time.Now()

	// Remove the current job after execution finished.
	e.jobs.Delete(job.ID)

	cb := CallbackParam{
		LogID:       job.LogID,
		LogDateTime: time.Now().UnixNano() / int64(time.Millisecond),
	}

	if err != nil {
		e.Logger.Error(logPrefix+"[%d:%d][%s] job handler failed: %v", job.ID, job.LogID, job.Duration(), err)
		cb.HandleCode = failureCode
		cb.HandleMsg = err.Error()
	} else {
		e.Logger.Info(logPrefix+"[%d:%d][%s] job handler succeeded", job.ID, job.LogID, job.Duration())
		cb.HandleCode = successCode
		cb.HandleMsg = "OK"
	}
	e.callbackChan <- cb
}

func (e *Executor) runJob(job *Job) {
	// wait for old job to finish before starting the new job
	st := time.Now()
	for {
		if oldJob := e.getJob(job.ID); oldJob == nil {
			break
		}
		time.Sleep(time.Second)
		e.Logger.Info(logPrefix+"[%d:%d] waiting for old job to finish, time elapsed: %s", job.ID, job.LogID, TruncateDuration(time.Since(st)))
	}
	e.addJob(job)
	e.Logger.Info(logPrefix+"[%d:%d] job starts", job.ID, job.LogID)
	job.Run()
}

/* Below are methods for serving http api */

func (e *Executor) setupRoutes() {
	http.HandleFunc("/beat", e.beat)
	http.HandleFunc("/idleBeat", e.idleBeat)
	http.HandleFunc("/run", e.trigger)
	http.HandleFunc("/kill", e.kill)
	http.HandleFunc("/log", e.log)
}

func (e *Executor) parseParam(r *http.Request, param interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return err
	}

	return json.Unmarshal(body, param)
}

// beat is for xxl-job server to check the executor's health.
func (e *Executor) beat(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	e.Logger.Info(logPrefix + "beat")

	fmt.Fprintln(w, NewSuccResponse().String())
}

// idleBeat is for xxl-job server to check if the job is idle.
func (e *Executor) idleBeat(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	var param IdleBeatParam
	if err := e.parseParam(r, &param); err != nil {
		fmt.Fprintln(w, NewErrorResponse(err.Error()).String())
		return
	}

	e.Logger.Info(logPrefix+"check idle of job %d", param.JobID)

	if e.getJob(param.JobID) != nil {
		fmt.Fprintln(w, NewErrorResponse("job is running").String())
		return
	}

	fmt.Fprintln(w, NewSuccResponse().String())
}

// trigger is for xxl-job server to trigger a job execution.
func (e *Executor) trigger(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	var params RunParam
	if err := e.parseParam(r, &params); err != nil {
		fmt.Fprintln(w, NewErrorResponse(err.Error()).String())
		return
	}

	if err := e.TriggerJob(params); err != nil {
		e.Logger.Error(logPrefix + fmt.Sprintf("[%d:%d] fail to trigger job: %v", params.JobID, params.LogID, err))
		fmt.Fprintln(w, NewErrorResponse(err.Error()))
		return
	}

	e.Logger.Info(logPrefix+"triger job: %+v", params)

	fmt.Fprintln(w, NewSuccResponse().String())
}

// kill is for xxl-job server to terminate a running job.
func (e *Executor) kill(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	var param KillParam
	if err := e.parseParam(r, &param); err != nil {
		fmt.Fprintln(w, NewErrorResponse(err.Error()).String())
		return
	}

	e.Logger.Info(logPrefix+"killing job %d", param.JobID)

	if job := e.getJob(param.JobID); job != nil {
		e.stopJob(job)
	}

	fmt.Fprintln(w, NewSuccResponse().String())
}

// log is for xxl-job server to retrieve job execution logs.
// Because we do not write any logs locally, we directly return a dummy response here.
func (e *Executor) log(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	var param LogParam
	if err := e.parseParam(r, &param); err != nil {
		fmt.Fprintln(w, NewErrorResponse(err.Error()).String())
		return
	}

	content := &LogResult{
		FromLineNum: 1,
		ToLineNum:   2,
		LogContent:  "N/A",
		IsEnd:       true,
	}

	res := NewSuccResponse()
	res.Content = content

	fmt.Fprintln(w, res.String())
}
