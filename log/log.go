package log

import (
	"fmt"
	"time"
)

type Logger interface {
	Debug(msg string, vars ...interface{})
	Info(msg string, vars ...interface{})
	Error(msg string, vars ...interface{})
}

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelError
)

const (
	levelDebugMsg = "Debug"
	levelInfoMsg  = "Info"
	levelErrorMsg = "Error"
)

type logger struct {
	level  Level
	logger Logger
}

func (l *logger) Debug(msg string, vars ...interface{}) {
	if l.logger != nil {
		l.logger.Debug(msg, vars...)
	} else if l.isOk(LevelDebug) {
		l.println(levelDebugMsg, msg, vars...)
	}
}

func (l *logger) Info(msg string, vars ...interface{}) {
	if l.logger != nil {
		l.logger.Info(msg, vars...)
	} else if l.isOk(LevelInfo) {
		l.println(levelInfoMsg, msg, vars...)
	}
}

func (l *logger) Error(msg string, vars ...interface{}) {
	if l.logger != nil {
		l.logger.Error(msg, vars...)
	} else if l.isOk(LevelError) {
		l.println(levelErrorMsg, msg, vars...)
	}
}

func (l *logger) println(level, msg string, vars ...interface{}) {
	msg = fmt.Sprintf("[%s][%s] %s\n", level, time.Now().String(), msg)
	fmt.Printf(msg, vars...)
}

func (l *logger) isOk(level Level) bool {
	return l.level <= level
}

func New(level Level, l Logger) Logger {
	return &logger{level: level, logger: l}
}
