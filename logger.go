package xxljob

import "log"

// Logger is a generic logger interface.
type Logger interface {
	Info(string, ...interface{})
	Error(string, ...interface{})
}

type dummyLogger struct{}

func (l dummyLogger) Info(string, ...interface{})  {}
func (l dummyLogger) Error(string, ...interface{}) {}

// DummyLogger returns a logger which writes nothing.
func DummyLogger() Logger {
	return new(dummyLogger)
}

type defaultLogger struct{}

func (l defaultLogger) Info(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func (l defaultLogger) Error(format string, v ...interface{}) {
	log.Printf(format, v...)
}

// DefaultLogger returns a default logger.
func DefaultLogger() Logger {
	return new(defaultLogger)
}
