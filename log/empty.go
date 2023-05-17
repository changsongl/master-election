package log

type emptyLogger struct {
}

func (e emptyLogger) Debug(msg string, vars ...interface{}) {
}

func (e emptyLogger) Info(msg string, vars ...interface{}) {
}

func (e emptyLogger) Error(msg string, vars ...interface{}) {
}

func NewEmptyLogger() Logger {
	return &emptyLogger{}
}
