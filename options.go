package xxljob

import (
	"os"
	"strings"
	"syscall"
	"time"
)

const (
	defaultCallbackBufferSize = 1024
	defaultCallbackInterval   = "1s"
	defaultClientTimeout      = time.Second * 3
	defaultRegisterInterval   = "10s"
	defaultSizeLimit          = 10240

	defaultPort        = 9999
	defaultIdleTimeout = time.Second * 60
	defaultReadTimeout = time.Second * 15
	defaultWrteTimeout = time.Second * 15
	defaultWaitTimeout = time.Second * 15
)

var defaultInterruptSignals = []os.Signal{
	syscall.SIGINT,
	syscall.SIGQUIT,
	syscall.SIGTERM,
}

// Options are executor options.
type Options struct {
	// client settings
	AccessToken        string
	AppName            string
	CallbackBufferSize int
	CallbackInterval   string
	ClientTimeout      time.Duration
	Host               string
	Logger             Logger
	RegisterInterval   string
	SizeLimit          int64 // we will not log the response if its size exceeds the size limit

	// http server settings
	Port             int
	IdleTimeout      time.Duration
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
	WaitTimeout      time.Duration
	interruptSignals []os.Signal
}

// NewOptions creates options with defaults
func NewOptions(opts ...Option) Options {
	var options = Options{
		CallbackBufferSize: defaultCallbackBufferSize,
		CallbackInterval:   defaultCallbackInterval,
		ClientTimeout:      defaultClientTimeout,
		Logger:             DefaultLogger(),
		RegisterInterval:   defaultRegisterInterval,
		SizeLimit:          defaultSizeLimit,

		Port:             defaultPort,
		IdleTimeout:      defaultIdleTimeout,
		ReadTimeout:      defaultReadTimeout,
		WriteTimeout:     defaultWrteTimeout,
		WaitTimeout:      defaultWaitTimeout,
		interruptSignals: defaultInterruptSignals,
	}

	for _, opt := range opts {
		opt(&options)
	}

	return options
}

// Option is for setting options.
type Option func(*Options)

// WithAccessToken sets access token.
func WithAccessToken(token string) Option {
	return func(o *Options) {
		o.AccessToken = token
	}
}

// WithAppName sets app name.
func WithAppName(appName string) Option {
	return func(o *Options) {
		o.AppName = appName
	}
}

// WithCallbackBufferSize sets callback buffer size.
func WithCallbackBufferSize(size int) Option {
	return func(o *Options) {
		o.CallbackBufferSize = size
	}
}

// WithCallbackInterval sets callback interval.
func WithCallbackInterval(interval string) Option {
	return func(o *Options) {
		o.CallbackInterval = interval
	}
}

// WithClientTimeout sets client timeout.
func WithClientTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.ClientTimeout = timeout
	}
}

// WithHost sets xxl-job server address.
func WithHost(host string) Option {
	return func(o *Options) {
		if !strings.HasPrefix(host, "http") {
			host = "http://" + host
		}
		o.Host = host
	}
}

// WithLogger sets logger.
func WithLogger(logger Logger) Option {
	return func(o *Options) {
		o.Logger = logger
	}
}

// WithRegisterInterval sets register interval.
func WithRegisterInterval(interval string) Option {
	return func(o *Options) {
		o.RegisterInterval = interval
	}
}

// WithSizeLimit sets size limit.
func WithSizeLimit(sizeLimit int64) Option {
	return func(o *Options) {
		o.SizeLimit = sizeLimit
	}
}

// WithPort sets service port.
func WithPort(port int) Option {
	return func(o *Options) {
		o.Port = port
	}
}

// WithIdleTimeout sets idle timeout.
func WithIdleTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.IdleTimeout = timeout
	}
}

// WithReadTimeout sets read timeout.
func WithReadTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.ReadTimeout = timeout
	}
}

// WithWriteTimeout sets write timeout.
func WithWriteTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.WriteTimeout = timeout
	}
}

// WithWaitTimeout sets wait timeout for graceful shutdown.
func WithWaitTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.WaitTimeout = timeout
	}
}

// WithInterruptSignals sets interrupt signals.
func WithInterruptSignals(signals []os.Signal) Option {
	return func(o *Options) {
		o.interruptSignals = signals
	}
}
