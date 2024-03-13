package logger

import (
	"log/slog"
	"os"
	"strings"
	"sync/atomic"
)

var stdLogger atomic.Pointer[Logger]
var errLogger atomic.Pointer[Logger]

type Logger struct {
	*slog.Logger
}

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

func NewLogLevel(level string) Level {
	level = strings.ToUpper(level)
	switch level {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	default:
		return INFO
	}
}

func getSLogLevel(level Level) slog.Level {
	switch level {
	case DEBUG:
		return slog.LevelDebug
	case INFO:
		return slog.LevelInfo
	case WARN:
		return slog.LevelWarn
	case ERROR:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

type ConfigOpt struct {
	Level Level
}

func (opt *ConfigOpt) WithLevel(level Level) *ConfigOpt {
	opt.Level = level
	return opt
}

func NewConfigOpt() *ConfigOpt {
	return &ConfigOpt{}
}

func initStdLogger(cfgOpt *ConfigOpt) {
	if cfgOpt == nil {
		cfgOpt = NewConfigOpt()
	}

	slogLevel := getSLogLevel(cfgOpt.Level)
	lvl := new(slog.LevelVar)
	lvl.Set(slogLevel)
	l := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: lvl,
	}))

	stdLogger.Store(&Logger{l})
}

func initErrLogger() {
	l := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	errLogger.Store(&Logger{l})
}

func Init(cfgOpt *ConfigOpt) {
	initStdLogger(cfgOpt)
	initErrLogger()
}

func Debug(msg string, args ...any) {
	stdLogger.Load().Debug(msg, args...)
}

func Info(msg string, args ...any) {
	stdLogger.Load().Info(msg, args...)
}

func Warn(msg string, args ...any) {
	stdLogger.Load().Warn(msg, args...)
}

func Error(msg string, args ...any) {
	errLogger.Load().Error(msg, args...)
}
