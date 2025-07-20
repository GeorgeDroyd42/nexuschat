package utils

import (
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

// Logger is the centralized application logger
var Logger = logrus.New()

// LogModule represents different modules in the application
type LogModule string

const (
	ModuleDatabase LogModule = "database"
	ModuleAuth     LogModule = "auth"
	ModuleUser     LogModule = "user"
	ModuleSession  LogModule = "session"
)

// Add these new variables for async logging
var logBuffer = make(chan *logEntry, 100) // Buffer size of 100 log entries
var logWaitGroup sync.WaitGroup

// logEntry represents a single log message to be processed
type logEntry struct {
	level     logrus.Level
	module    LogModule
	operation string
	message   string
	err       error
}

func InitLogger() {
	Logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	Logger.SetLevel(logrus.InfoLevel)

	Logger.SetOutput(os.Stdout)

	go processLogs()
}

func processLogs() {
	for entry := range logBuffer {
		fields := logrus.Fields{
			"module":    entry.module,
			"operation": entry.operation,
		}

		switch entry.level {
		case logrus.ErrorLevel:
			Logger.WithFields(fields).Errorf("Error: %v", entry.err)
		case logrus.InfoLevel:
			Logger.WithFields(fields).Info(entry.message)
		case logrus.DebugLevel:
			Logger.WithFields(fields).Debug(entry.message)
		}

		logWaitGroup.Done()
	}
}
func log(level logrus.Level, module LogModule, operation string, message string, err error) {
	logWaitGroup.Add(1)
	select {
	case logBuffer <- &logEntry{
		level:     level,
		module:    module,
		operation: operation,
		message:   message,
		err:       err,
	}:
		// Successfully queued
	default:
		// Buffer full, log synchronously as fallback
		fields := logrus.Fields{
			"module":    module,
			"operation": operation,
		}

		switch level {
		case logrus.ErrorLevel:
			Logger.WithFields(fields).Errorf("Error: %v", err)
		case logrus.InfoLevel:
			Logger.WithFields(fields).Info(message)
		case logrus.DebugLevel:
			Logger.WithFields(fields).Debug(message)
		}

		logWaitGroup.Done()
	}
}
func FlushLogs() {
	logWaitGroup.Wait()
}

func Log(level logrus.Level, module string, operation string, message string, err error) {
	log(level, LogModule(module), operation, message, err)
}
