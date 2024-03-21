# xxljob

[![GoDoc](https://pkg.go.dev/badge/github.com/hyperjiang/xxljob)](https://pkg.go.dev/github.com/hyperjiang/xxljob)
[![CI](https://github.com/hyperjiang/xxljob/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/hyperjiang/xxljob/actions/workflows/ci.yml)
[![](https://goreportcard.com/badge/github.com/hyperjiang/xxljob)](https://goreportcard.com/report/github.com/hyperjiang/xxljob)
[![codecov](https://codecov.io/gh/hyperjiang/xxljob/branch/main/graph/badge.svg)](https://codecov.io/gh/hyperjiang/xxljob)
[![Release](https://img.shields.io/github/release/hyperjiang/xxljob.svg)](https://github.com/hyperjiang/xxljob/releases)

XXL-JOB golang executor is a standalone http server which manages the connection with XXL-JOB server.

The default port is 9999. It registers itself as a node in XXL-JOB server, listens and handles the calls from XXL-JOB server.

A golang executor can run multiple jobs concurrently, but each job is run in serial mode, which is strictly following the design pattern of XXL-JOB:

> XXL-JOB的不同任务之间并行调度、并行执行。
> XXL-JOB的单个任务，针对多个执行器是并行运行的，**针对单个执行器是串行执行的**。同时支持任务终止。

See [5.4.5 并行调度](https://github.com/xuxueli/xxl-job/blob/72963e4716a74eacdcbdd2e999c433debf3afaa3/doc/XXL-JOB%E5%AE%98%E6%96%B9%E6%96%87%E6%A1%A3.md#545-%E5%B9%B6%E8%A1%8C%E8%B0%83%E5%BA%A6)

## Prerequisite

go version >= 1.16

## Installation

```
go get -u github.com/hyperjiang/xxljob
```

## Usage

### 1. Start the executor

```go
import "github.com/hyperjiang/xxljob"

const (
	appName     = "xxl-job-executor-sample"
	accessToken = "default_token"
	host        = "localhost:8080/xxl-job-admin"
	demoHandler = "demoJobHandler"
)

e := xxljob.NewExecutor(
    xxljob.WithAppName(appName),
    xxljob.WithAccessToken(accessToken),
    xxljob.WithHost(host),
)

// start in goroutine
go e.Start()
```
### 2. Add job handler

Job handlers are functions that implement `xxljob.JobHandler` (that is `func(ctx context.Context, param xxljob.JobParam) error`).

```go
e.AddJobHandler(demoHandler, func(ctx context.Context, param xxljob.JobParam) error {
    fmt.Println(param.Params)
    return nil
})
```

### 3. Stop the executor

```go
e.Stop()
```
