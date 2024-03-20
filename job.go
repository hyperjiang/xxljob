package xxljob

import (
	"context"
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
	ID        int
	LogID     int64
	Name      string
	Handle    JobHandler
	Param     JobParam
	Timeout   int // timeout in seconds
	StartTime time.Time
	EndTime   time.Time

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
	if j.Timeout > 0 {
		j.ctx, j.cancel = context.WithTimeout(j.ctx, time.Duration(j.Timeout)*time.Second)
	} else {
		j.ctx, j.cancel = context.WithCancel(j.ctx)
	}

	j.StartTime = time.Now()
	j.done <- j.Handle(j.ctx, j.Param)
}

// Stop stops the job.
func (j *Job) Stop() {
	j.cancel()
}

// Duration returns the duration of job execution.
func (j *Job) Duration() time.Duration {
	return TruncateDuration(j.EndTime.Sub(j.StartTime))
}
