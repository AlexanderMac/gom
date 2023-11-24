package gom

import "log"

const (
	LOG_LEVEL_DEBUG = iota
	LOG_LEVEL_INFO
	LOG_LEVEL_WARN
	LOG_LEVEL_ERROR
)

type Logger interface {
	Debug(args ...any)
	Info(args ...any)
	Warn(args ...any)
	Error(args ...any)

	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
}

type defLogger struct {
	logLevel int
}

var logger Logger = &defLogger{
	logLevel: LOG_LEVEL_INFO,
}

func SetLogger(l Logger) {
	logger = l
}

func SetDefLogLevel(logLevel int) {
	logger = &defLogger{
		logLevel,
	}
}

func (l *defLogger) Debug(args ...any) {
	if l.logLevel <= LOG_LEVEL_DEBUG {
		log.Println(args...)
	}
}

func (l *defLogger) Info(args ...any) {
	if l.logLevel <= LOG_LEVEL_INFO {
		log.Println(args...)
	}
}

func (l *defLogger) Warn(args ...any) {
	if l.logLevel <= LOG_LEVEL_WARN {
		log.Println(args...)
	}
}

func (l *defLogger) Error(args ...any) {
	if l.logLevel <= LOG_LEVEL_ERROR {
		log.Println(args...)
	}
}

func (l *defLogger) Debugf(format string, args ...any) {
	if l.logLevel <= LOG_LEVEL_DEBUG {
		log.Printf(format, args...)
	}
}

func (l *defLogger) Infof(format string, args ...any) {
	if l.logLevel <= LOG_LEVEL_INFO {
		log.Printf(format, args...)
	}
}

func (l *defLogger) Warnf(format string, args ...any) {
	if l.logLevel <= LOG_LEVEL_WARN {
		log.Printf(format, args...)
	}
}

func (l *defLogger) Errorf(format string, args ...any) {
	if l.logLevel <= LOG_LEVEL_ERROR {
		log.Printf(format, args...)
	}
}
