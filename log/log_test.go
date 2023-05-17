package log

import (
	"os"
	"testing"
)

func TestLogger(t *testing.T) {
	l := New(LevelInfo, nil)
	l.Debugf("this a debug msg")
	l.Infof("this a info msg")
	l.Errorf("this a error msg")

	l.Infof(os.Getenv("username"))
}
