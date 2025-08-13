package core

import (
	"log"
	"os"
)

// Logger interface for connection pool logging
type Logger interface {
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
}

// DefaultLogger provides a simple logger implementation
type DefaultLogger struct {
	logger *log.Logger
}

// NewDefaultLogger creates a new default logger
func NewDefaultLogger() *DefaultLogger {
	return &DefaultLogger{
		logger: log.New(os.Stdout, "[ConnectionPool] ", log.LstdFlags),
	}
}

func (l *DefaultLogger) Info(msg string, args ...interface{}) {
	l.logger.Printf("INFO: "+msg, args...)
}

func (l *DefaultLogger) Warn(msg string, args ...interface{}) {
	l.logger.Printf("WARN: "+msg, args...)
}

func (l *DefaultLogger) Error(msg string, args ...interface{}) {
	l.logger.Printf("ERROR: "+msg, args...)
}

func (l *DefaultLogger) Debug(msg string, args ...interface{}) {
	l.logger.Printf("DEBUG: "+msg, args...)
}