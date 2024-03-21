package xxljob_test

import (
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/hyperjiang/xxljob"
	"github.com/stretchr/testify/require"
)

func TestOptions(t *testing.T) {
	should := require.New(t)

	// default options
	opts := xxljob.NewOptions()
	should.Empty(opts.AccessToken)
	should.Equal(1024, opts.CallbackBufferSize)
	should.Equal("1s", opts.CallbackInterval)
	should.Equal(time.Second*3, opts.ClientTimeout)
	should.Empty(opts.Host)
	should.Equal("10s", opts.RegisterInterval)
	should.Equal(int64(10240), opts.SizeLimit)

	should.Equal(9999, opts.Port)
	should.Equal(time.Second*60, opts.IdleTimeout)
	should.Equal(time.Second*15, opts.ReadTimeout)
	should.Equal(time.Second*15, opts.WriteTimeout)
	should.Equal(time.Second*15, opts.WaitTimeout)

	// override default options
	opts2 := xxljob.NewOptions(
		xxljob.WithAccessToken("abc"),
		xxljob.WithAppName(appName),
		xxljob.WithCallbackBufferSize(100),
		xxljob.WithCallbackInterval("5s"),
		xxljob.WithClientTimeout(time.Second),
		xxljob.WithHost(host),
		xxljob.WithLogger(xxljob.DummyLogger()),
		xxljob.WithRegisterInterval("15s"),
		xxljob.WithSizeLimit(20000),

		xxljob.WithPort(8080),
		xxljob.WithIdleTimeout(time.Second*10),
		xxljob.WithReadTimeout(time.Second),
		xxljob.WithWriteTimeout(time.Second*2),
		xxljob.WithWaitTimeout(time.Second*3),
		xxljob.WithInterruptSignals([]os.Signal{syscall.SIGKILL}),
	)

	should.Equal("abc", opts2.AccessToken)
	should.Equal(appName, opts2.AppName)
	should.Equal(100, opts2.CallbackBufferSize)
	should.Equal("5s", opts2.CallbackInterval)
	should.Equal(time.Second, opts2.ClientTimeout)
	should.Equal("http://"+host, opts2.Host)
	should.Equal("15s", opts2.RegisterInterval)
	should.Equal(int64(20000), opts2.SizeLimit)

	should.Equal(8080, opts2.Port)
	should.Equal(time.Second*10, opts2.IdleTimeout)
	should.Equal(time.Second, opts2.ReadTimeout)
	should.Equal(time.Second*2, opts2.WriteTimeout)
	should.Equal(time.Second*3, opts2.WaitTimeout)
}
