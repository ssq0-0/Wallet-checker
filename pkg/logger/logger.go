package logger

import (
	"io"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
)

// Package logger provides a logging interface and implementation using logrus.
// It supports different log levels and structured logging with fields.

// GlobalLogger is a singleton instance of the logger that can be used throughout the application.
var GlobalLogger Logger

func init() {
	GlobalLogger = New(InfoLevel)
}

func Init(level LoggerLevel) {
	GlobalLogger = New(level)
}

// ParseLevel преобразует строковое представление уровня логирования в LoggerLevel
func ParseLevel(levelStr string) LoggerLevel {
	levelStr = strings.ToLower(levelStr)
	switch levelStr {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn", "warning":
		return WarnLevel
	case "error":
		return ErrorLevel
	case "fatal":
		return FatalLevel
	case "panic":
		return PanicLevel
	default:
		return InfoLevel // По умолчанию используем InfoLevel
	}
}

// Fields represents a map of key-value pairs that can be added to log entries.
type Fields map[string]interface{}

// LoggerLevel represents the severity level of log messages.
type LoggerLevel string

const (
	// DebugLevel represents debug level logging.
	DebugLevel LoggerLevel = "debug"

	// InfoLevel represents informational level logging.
	InfoLevel LoggerLevel = "info"

	// WarnLevel represents warning level logging.
	WarnLevel LoggerLevel = "warn"

	// ErrorLevel represents error level logging.
	ErrorLevel LoggerLevel = "error"

	// FatalLevel represents fatal level logging.
	FatalLevel LoggerLevel = "fatal"

	// PanicLevel represents panic level logging.
	PanicLevel LoggerLevel = "panic"
)

// Logger defines the interface for logging operations.
// It provides methods for logging at different severity levels
// and supports structured logging with fields.
type Logger interface {
	// Debug logs a message at debug level.
	Debug(args ...interface{})

	// Debugf logs a formatted message at debug level.
	Debugf(format string, args ...interface{})

	// Error logs a message at error level.
	Error(args ...interface{})

	// Errorf logs a formatted message at error level.
	Errorf(format string, args ...interface{})

	// Fatal logs a message at fatal level and exits the program.
	Fatal(args ...interface{})

	// Fatalf logs a formatted message at fatal level and exits the program.
	Fatalf(format string, args ...interface{})

	// Info logs a message at info level.
	Info(args ...interface{})

	// Infof logs a formatted message at info level.
	Infof(format string, args ...interface{})

	// Panic logs a message at panic level and panics.
	Panic(args ...interface{})

	// Panicf logs a formatted message at panic level and panics.
	Panicf(format string, args ...interface{})

	// Warn logs a message at warn level.
	Warn(args ...interface{})

	// Warnf logs a formatted message at warn level.
	Warnf(format string, args ...interface{})

	// WithFields returns a new logger instance with the specified fields.
	WithFields(fields Fields) Logger
}

// logger implements the Logger interface using logrus.
type logger struct {
	*logrus.Logger
}

// New creates a new logger instance with the specified log level.
// The logger is configured to output to stdout with color support
// and includes timestamps in the log messages.
func New(level LoggerLevel) Logger {
	logrusLevel, err := logrus.ParseLevel(string(level))
	if err != nil {
		logrusLevel = logrus.InfoLevel
	}

	lgr := logrus.New()
	lgr.SetLevel(logrusLevel)
	lgr.SetOutput(io.Discard)

	// Настраиваем форматтер (например, JSON)
	lgr.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})

	// Добавляем асинхронный хук для записи в os.Stdout
	lgr.AddHook(&writer.Hook{
		Writer:    os.Stdout,
		LogLevels: logrus.AllLevels,
	})

	return logger{lgr}
}

// WithFields returns a new logger instance with the specified fields.
// The fields will be included in all subsequent log messages.
func (l logger) WithFields(fields Fields) Logger {
	return logger{
		Logger: l.Logger.WithFields(logrus.Fields(fields)).Logger,
	}
}
