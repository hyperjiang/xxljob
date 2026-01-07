package main

import (
	"context"
	"fmt"
	"time"

	"github.com/hyperjiang/xxljob"
)

const (
	appName     = "xxl-job-executor-sample"
	accessToken = "default_token"
	host        = "localhost:8080/xxl-job-admin"
	demoHandler = "demoJobHandler"
)

func main() {
	e := xxljob.NewExecutor(
		xxljob.WithAppName(appName),
		xxljob.WithAccessToken(accessToken),
		xxljob.WithClientTimeout(time.Second),
		xxljob.WithHost(host),
	)

	e.AddJobHandler(demoHandler, func(ctx context.Context, param xxljob.JobParam) error {
		fmt.Println(param.Params)
		logger := xxljob.LoggerFromContext(ctx)
		logger.Info("Job executed with params: %s", param.Params)
		return nil
	})

	if err := e.Start(); err != nil {
		panic(err)
	}
}
