package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Debugf(msg string, vars ...interface{})
	Infof(msg string, vars ...interface{})
	Errorf(msg string, vars ...interface{})
}

type Level zapcore.Level

const (
	LevelDebug Level = Level(zapcore.DebugLevel)
	LevelInfo  Level = Level(zapcore.InfoLevel)
	LevelError Level = Level(zapcore.ErrorLevel)
)

type logger struct {
	level  Level
	logger Logger
}

func (l *logger) Debugf(format string, vars ...interface{}) {
	l.logger.Debugf(format, vars...)
}

func (l *logger) Infof(format string, vars ...interface{}) {
	l.logger.Infof(format, vars...)
}

func (l *logger) Errorf(format string, vars ...interface{}) {
	l.logger.Errorf(format, vars...)
}

func New(level Level, l Logger) Logger {
	if l == nil {
		conf := zap.NewProductionConfig()
		conf.Encoding = "console"
		conf.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

		zapLevel := zap.NewAtomicLevel()
		zapLevel.SetLevel(zapcore.Level(level))
		conf.Level = zapLevel

		lg, _ := conf.Build(zap.AddCallerSkip(1))
		l = lg.Sugar()
	}

	return &logger{logger: l}
}
