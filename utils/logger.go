package utils

import (
	"github.com/sirupsen/logrus"
)

// LogLevel represents different log levels
type LogLevel int

const (
	LogError LogLevel = iota
	LogInfo
	LogDebug
	LogWarn
)

// Logger provides centralized structured logging
type Logger struct{}

var Log = &Logger{}

// LogWithFields logs with module, action and optional additional fields
func (l *Logger) LogWithFields(level LogLevel, module, action, message string, err error, fields map[string]interface{}) {
	logFields := logrus.Fields{
		"module": module,
		"action": action,
	}
	
	// Add additional fields
	for k, v := range fields {
		logFields[k] = v
	}
	
	entry := logrus.WithFields(logFields)
	if err != nil {
		entry = entry.WithError(err)
	}
	
	switch level {
	case LogError:
		entry.Error(message)
	case LogInfo:
		entry.Info(message)
	case LogDebug:
		entry.Debug(message)
	case LogWarn:
		entry.Warn(message)
	}
}

// Convenience methods
func (l *Logger) Error(module, action, message string, err error, fields ...map[string]interface{}) {
	f := make(map[string]interface{})
	if len(fields) > 0 {
		f = fields[0]
	}
	l.LogWithFields(LogError, module, action, message, err, f)
}

func (l *Logger) Info(module, action, message string, fields ...map[string]interface{}) {
	f := make(map[string]interface{})
	if len(fields) > 0 {
		f = fields[0]
	}
	l.LogWithFields(LogInfo, module, action, message, nil, f)
}

func (l *Logger) Debug(module, action, message string, fields ...map[string]interface{}) {
	f := make(map[string]interface{})
	if len(fields) > 0 {
		f = fields[0]
	}
	l.LogWithFields(LogDebug, module, action, message, nil, f)
}