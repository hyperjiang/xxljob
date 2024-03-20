package xxljob

/* Below are parameters of the local client to call the remote server */

// RegistryParam is to register or deregister an executor.
type RegistryParam struct {
	RegistryGroup string `json:"registryGroup"`
	RegistryKey   string `json:"registryKey"`
	RegistryValue string `json:"registryValue"`
}

// CallbackParam is to report job execution result.
type CallbackParam struct {
	LogID       int64  `json:"logId"`
	LogDateTime int64  `json:"logDateTim"`
	HandleCode  int    `json:"handleCode"`
	HandleMsg   string `json:"handleMsg"`
}

/* Below are the parameters of the remote server to call the local client's HTTP endpoints */

// IdleBeatParam is for idle checking.
type IdleBeatParam struct {
	JobID int `json:"jobId"`
}

// KillParam is used to terminate a running job.
type KillParam struct {
	JobID int `json:"jobId"`
}

// LogParam is used to get job logs.
type LogParam struct {
	LogId       int64 `json:"logId"`
	LogDateTime int64 `json:"logDateTim"`
	FromLineNum int   `json:"fromLineNum"`
}

// RunParam is used to trigger a job.
type RunParam struct {
	JobID                 int    `json:"jobId"`
	ExecutorHandler       string `json:"executorHandler"`
	ExecutorParams        string `json:"executorParams"`
	ExecutorBlockStrategy string `json:"executorBlockStrategy"`
	ExecutorTimeout       int    `json:"executorTimeout"` // job execution timeout in seconds
	LogID                 int64  `json:"logId"`
	LogDateTime           int64  `json:"logDateTime"`    // timestamp in milliseconds
	GlueType              string `json:"glueType"`       // unsupported in go executor
	GlueSource            string `json:"glueSource"`     // unsupported in go executor
	GlueUpdatetime        int64  `json:"glueUpdatetime"` // unsupported in go executor
	BroadcastIndex        int    `json:"broadcastIndex"`
	BroadcastTotal        int    `json:"broadcastTotal"`
}
