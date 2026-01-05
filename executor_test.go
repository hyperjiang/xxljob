package xxljob_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/hyperjiang/xxljob"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	appName     = "xxl-job-executor-sample"
	accessToken = "default_token"
	host        = "localhost:8080/xxl-job-admin"
	demoHandler = "demoJobHandler"
)

type ExecutorTestSuite struct {
	suite.Suite
	e *xxljob.Executor
}

func timestampMS() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// TestExecutorTestSuite runs the ljob client test suite.
func TestExecutorTestSuite(t *testing.T) {
	suite.Run(t, new(ExecutorTestSuite))
}

// SetupSuite run once at the very start of the testing suite, before any tests are run.
func (ts *ExecutorTestSuite) SetupSuite() {
	should := require.New(ts.T())
	ts.e = xxljob.NewExecutor(
		xxljob.WithAppName(appName),
		xxljob.WithAccessToken(accessToken),
		xxljob.WithClientTimeout(time.Second),
		xxljob.WithHost(host),
		xxljob.WithLogRetentionDays(1),
		xxljob.WithLogCleanupInterval("1s"),
		xxljob.WithLogDir("/tmp/xxl-job/jobhandler"),
	)

	ts.e.AddJobHandler(demoHandler, func(ctx context.Context, param xxljob.JobParam) error {
		fmt.Println(param.Params)
		return nil
	})

	ts.e.AddJobHandler("delayHandler", func(ctx context.Context, param xxljob.JobParam) error {
		// time.Sleep(time.Second * 5)
		for i := 0; i < 5; i++ {
			select {
			case <-ctx.Done():
				// context timeout or cancelled
				fmt.Println("context cancelled or timed out")
				return ctx.Err()
			default:
				fmt.Println(i)
				time.Sleep(time.Second)
			}
		}
		return nil
	})

	go func() {
		err := ts.e.Start()
		should.NoError(err)
	}()

	// wait for server start
	time.Sleep(time.Second)
}

// TearDownSuite run once at the very end of the testing suite, after all tests have been run
func (ts *ExecutorTestSuite) TearDownSuite() {
	time.Sleep(time.Second)
	ts.e.RemoveJobHandler(demoHandler)
	ts.e.RemoveJobHandler("delayHandler")
	_ = ts.e.Stop()
}

func (ts *ExecutorTestSuite) TestHappyPath() {
	should := require.New(ts.T())

	cli := resty.New().SetBaseURL(fmt.Sprintf("http://localhost:%d", ts.e.Port))

	resp, err := cli.R().Get("beat")
	should.NoError(err)

	var res xxljob.Response
	err = json.Unmarshal(resp.Body(), &res)
	should.NoError(err)
	should.Equal(200, res.Code)

	req := xxljob.IdleBeatParam{JobID: 1}
	resp, err = cli.R().SetBody(req).Post("idleBeat")
	should.NoError(err)
	err = json.Unmarshal(resp.Body(), &res)
	should.NoError(err)
	should.Equal(200, res.Code)

	req2 := xxljob.LogParam{LogId: 1}
	resp, err = cli.R().SetBody(req2).Post("log")
	should.NoError(err)
	err = json.Unmarshal(resp.Body(), &res)
	should.NoError(err)
	should.Equal(200, res.Code)

	req3 := xxljob.RunParam{
		JobID:           1,
		ExecutorHandler: demoHandler,
		ExecutorParams:  "this is demo handler params",
		LogID:           1,
		LogDateTime:     timestampMS(),
	}
	resp, err = cli.R().SetBody(req3).Post("run")
	should.NoError(err)
	err = json.Unmarshal(resp.Body(), &res)
	should.NoError(err)
	should.Equal(200, res.Code)

	// try to trigger a unexisting job handler, should fail
	req3 = xxljob.RunParam{
		JobID:           2,
		ExecutorHandler: "unexistingHandler",
		LogID:           2,
		LogDateTime:     timestampMS(),
	}
	resp, err = cli.R().SetBody(req3).Post("run")
	should.NoError(err)
	err = json.Unmarshal(resp.Body(), &res)
	should.NoError(err)
	should.Equal(500, res.Code)

	req3 = xxljob.RunParam{
		JobID:           2,
		ExecutorHandler: "delayHandler",
		ExecutorTimeout: 1,
		LogID:           2,
		LogDateTime:     timestampMS(),
	}
	resp, err = cli.R().SetBody(req3).Post("run")
	should.NoError(err)
	err = json.Unmarshal(resp.Body(), &res)
	should.NoError(err)
	should.Equal(200, res.Code)

	// duplicate log id
	resp, err = cli.R().SetBody(req3).Post("run")
	should.NoError(err)
	err = json.Unmarshal(resp.Body(), &res)
	should.NoError(err)
	should.Equal(500, res.Code)

	// the job is still running, this request is discarded
	req3.ExecutorBlockStrategy = xxljob.DiscardLater
	req3.LogID = 3
	resp, err = cli.R().SetBody(req3).Post("run")
	should.NoError(err)
	err = json.Unmarshal(resp.Body(), &res)
	should.NoError(err)
	should.Equal(500, res.Code)

	// terminate the running worker and clear the worker queue,
	// and then run the new request
	req3.ExecutorBlockStrategy = xxljob.CoverEarly
	req3.LogID = 4
	resp, err = cli.R().SetBody(req3).Post("run")
	should.NoError(err)
	err = json.Unmarshal(resp.Body(), &res)
	should.NoError(err)
	should.Equal(200, res.Code)
	// time.Sleep(time.Second) // wait for context timeout

	req3.ExecutorBlockStrategy = xxljob.SerialExecution
	req3.LogID = 5
	req3.ExecutorTimeout = 10
	resp, err = cli.R().SetBody(req3).Post("run")
	should.NoError(err)
	err = json.Unmarshal(resp.Body(), &res)
	should.NoError(err)
	should.Equal(200, res.Code)

	// the job is running (not idle)
	req = xxljob.IdleBeatParam{JobID: 2}
	resp, err = cli.R().SetBody(req).Post("idleBeat")
	should.NoError(err)
	err = json.Unmarshal(resp.Body(), &res)
	should.NoError(err)
	should.Equal(500, res.Code)
	fmt.Println(res.Msg)

	req4 := xxljob.KillParam{
		JobID: 2,
	}
	resp, err = cli.R().SetBody(req4).Post("kill")
	should.NoError(err)
	err = json.Unmarshal(resp.Body(), &res)
	should.NoError(err)
	should.Equal(200, res.Code)
}

func (ts *ExecutorTestSuite) TestInvalidRequest() {
	should := require.New(ts.T())

	cli := resty.New().SetBaseURL(fmt.Sprintf("http://localhost:%d", ts.e.Port))

	var res xxljob.Response

	endpoints := []string{"idleBeat", "log", "run", "kill"}
	for _, ep := range endpoints {
		resp, err := cli.R().SetBody("invalid input").Post(ep)
		should.NoError(err)
		err = json.Unmarshal(resp.Body(), &res)
		should.NoError(err)
		should.Equal(500, res.Code)
	}
}
