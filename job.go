package xxljob

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	// SerialExecution: the scheduling job enters the FIFO queue and runs in serial mode. (default)
	SerialExecution = "SERIAL_EXECUTION"
	// DiscardLater: if there are running jobs in the executor,
	// this job will be discarded and marked as failed.
	DiscardLater = "DISCARD_LATER"
	// CoverEarly: if there are running jobs in the executor,
	// the running jobs will be terminated and the queue will be cleared,
	// and then this new job will be run.
	CoverEarly = "COVER_EARLY"
)

// Job represents a scheduled job.
type Job struct {
	ID          int
	LogID       int64
	LogDateTime int64
	Name        string
	Handle      JobHandler
	Param       JobParam
	Timeout     int // timeout in seconds
	StartTime   time.Time
	EndTime     time.Time
	LogDir      string

	ctx    context.Context
	cancel context.CancelFunc
	done   chan error
}

// JobParam is the parameter passed to the job handler.
type JobParam struct {
	Params        string
	ShardingIndex int
	ShardingTotal int
}

// JobHandler is the handler function for executing job.
type JobHandler func(ctx context.Context, param JobParam) error

// Run runs the job.
func (j *Job) Run() {
	if j.ctx == nil {
		j.ctx = context.Background()
	}

	if j.Timeout > 0 {
		j.ctx, j.cancel = context.WithTimeout(j.ctx, time.Duration(j.Timeout)*time.Second)
	} else {
		j.ctx, j.cancel = context.WithCancel(j.ctx)
	}

	var jobLogger *fileLogger

	// Prepare log file if LogDir is configured.
	if j.LogDir != "" {
		logDate := time.Unix(j.LogDateTime/1000, 0)
		logDir := filepath.Join(j.LogDir, logDate.Format("2006-01-02"))
		if err := os.MkdirAll(logDir, 0755); err == nil {
			logFile := filepath.Join(logDir, fmt.Sprintf("%d.log", j.LogID))
			if f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
				logger := &fileLogger{file: f}
				jobLogger = logger
				j.ctx = ContextWithLogger(j.ctx, logger)
				defer logger.Close()
			}
		}
	}

	j.StartTime = time.Now()
	if jobLogger != nil {
		jobLogger.Info("job start: id=%d logId=%d handler=%s params=%s", j.ID, j.LogID, j.Name, j.Param.Params)
	}

	err := j.Handle(j.ctx, j.Param)

	if jobLogger != nil {
		if err != nil {
			jobLogger.Error("job failed: %v", err)
		} else {
			jobLogger.Info("job success")
		}
	}

	j.done <- err
}

// Stop stops the job.
func (j *Job) Stop() {
	j.cancel()
}

// Duration returns the duration of job execution.
func (j *Job) Duration() time.Duration {
	return TruncateDuration(j.EndTime.Sub(j.StartTime))
}
