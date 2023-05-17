package master

import (
	"fmt"
	"time"
)

type Logger interface {
	Debug(msg string, vars ...interface{})
	Info(msg string, vars ...interface{})
	Error(msg string, vars ...interface{})
}

type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelError
)

const (
	levelDebugMsg = "Debug"
	levelInfoMsg  = "Info"
	levelErrorMsg = "Error"
)

type logger struct {
	level  LogLevel
	logger Logger
}

func (l *logger) Debug(msg string, vars ...interface{}) {
	if !l.isOk(LogLevelDebug) {
		return
	}

	if l.logger != nil {
		l.logger.Debug(msg, vars...)
	} else {
		l.println(levelDebugMsg, msg, vars...)
	}
}

func (l *logger) Info(msg string, vars ...interface{}) {
	if !l.isOk(LogLevelInfo) {
		return
	}

	if l.logger != nil {
		l.logger.Info(msg, vars...)
	} else {
		l.println(levelInfoMsg, msg, vars...)
	}
}

func (l *logger) Error(msg string, vars ...interface{}) {
	if !l.isOk(LogLevelError) {
		return
	}

	if l.logger != nil {
		l.logger.Error(msg, vars...)
	} else {
		l.println(levelErrorMsg, msg, vars...)
	}
}

func (l *logger) println(level, msg string, vars ...interface{}) {
	msg = fmt.Sprintf("[%s][%s] %s\n", level, time.Now().String(), msg)
	fmt.Printf(msg, vars...)
}

func (l *logger) isOk(level LogLevel) bool {
	return l.level <= level
}

func newLogger(level LogLevel, l Logger) Logger {
	return &logger{level: level, logger: l}
}
