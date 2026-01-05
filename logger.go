package xxljob

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
)

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

type contextKey string

const jobLoggerKey contextKey = "xxljob_job_logger"

// ContextWithLogger puts the logger into context so handlers can retrieve it.
func ContextWithLogger(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, jobLoggerKey, logger)
}

type fileLogger struct {
	file *os.File
}

func (l *fileLogger) Info(format string, v ...interface{}) {
	l.write("INFO", format, v...)
}

func (l *fileLogger) Error(format string, v ...interface{}) {
	l.write("ERROR", format, v...)
}

func (l *fileLogger) write(level, format string, v ...interface{}) {
	if l.file == nil {
		return
	}
	msg := fmt.Sprintf(format, v...)
	now := time.Now().Format("2006-01-02 15:04:05")
	line := fmt.Sprintf("%s [%s] %s\n", now, level, msg)
	_, _ = l.file.WriteString(line)
}

func (l *fileLogger) Close() {
	if l.file != nil {
		_ = l.file.Close()
	}
}
