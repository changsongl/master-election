package master

import "testing"

func TestLogger(t *testing.T) {
	l := newLogger(LogLevelInfo, nil)
	l.Debug("this a debug msg")
	l.Info("this a info msg")
	l.Error("this a error msg")
}
