// Package log is the logging library used by IPFS & libp2p
// (https://github.com/ipfs/go-ipfs).
package log

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// StandardLogger provides API compatibility with standard printf loggers
// eg. go-logging
type StandardLogger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Panic(args ...interface{})
	Panicf(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
}

// EventLogger extends the StandardLogger interface to allow for log items
// containing structured metadata
type EventLogger interface {
	StandardLogger
}

// Logger retrieves an event logger by name
func Logger(system string) *ZapEventLogger {
	if len(system) == 0 {
		setuplog := getLogger("setup-logger")
		setuplog.Error("Missing name parameter")
		system = "undefined"
	}

	logger := getLogger(system)
	skipLogger := logger.Desugar().WithOptions(zap.AddCallerSkip(1)).Sugar()

	return &ZapEventLogger{
		system:        system,
		SugaredLogger: *logger,
		skipLogger:    *skipLogger,
	}
}

// ZapEventLogger implements the EventLogger and wraps a go-logging Logger
type ZapEventLogger struct {
	zap.SugaredLogger
	// used to fix the caller location when calling Warning and Warningf.
	skipLogger zap.SugaredLogger
	system     string
}

// Warning is for compatibility
// Deprecated: use Warn(args ...interface{}) instead
func (logger *ZapEventLogger) Warning(args ...interface{}) {
	logger.skipLogger.Warn(args...)
}

// Warningf is for compatibility
// Deprecated: use Warnf(format string, args ...interface{}) instead
func (logger *ZapEventLogger) Warningf(format string, args ...interface{}) {
	logger.skipLogger.Warnf(format, args...)
}

// FormatRFC3339 returns the given time in UTC with RFC3999Nano format.
func FormatRFC3339(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func WithStacktrace(l *ZapEventLogger, level LogLevel) *ZapEventLogger {
	copyLogger := *l
	copyLogger.SugaredLogger = *copyLogger.SugaredLogger.Desugar().
		WithOptions(zap.AddStacktrace(zapcore.Level(level))).Sugar()
	copyLogger.skipLogger = *copyLogger.SugaredLogger.Desugar().WithOptions(zap.AddCallerSkip(1)).Sugar()
	return &copyLogger
}

// WithSkip returns a new logger that skips the specified number of stack frames when reporting the
// line/file.
func WithSkip(l *ZapEventLogger, skip int) *ZapEventLogger {
	copyLogger := *l
	copyLogger.SugaredLogger = *copyLogger.SugaredLogger.Desugar().
		WithOptions(zap.AddCallerSkip(skip)).Sugar()
	copyLogger.skipLogger = *copyLogger.SugaredLogger.Desugar().WithOptions(zap.AddCallerSkip(1)).Sugar()
	return &copyLogger
}

func NewLoggerWithLevel(name string, level string) *ZapEventLogger {
	log := Logger(name)
	cfg := configFromEnv()
	if l, ok := cfg.SubsystemLevels[name]; ok {
		levels[name].SetLevel(zapcore.Level(l))
	} else {
		if err := SetLogLevel(name, level); err != nil {
			panic(err)
		}
	}

	return log
}

// return an logger with info level
func NewLogger(name string) *ZapEventLogger {
	l := Logger(name)
	cfg := configFromEnv()
	if lvl, ok := cfg.SubsystemLevels[name]; ok {
		levels[name].SetLevel(zapcore.Level(lvl))
	} else {
		SetLogLevel(name, "INFO")
	}
	return l
}
