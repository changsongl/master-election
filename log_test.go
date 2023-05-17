package master

import (
	"github.com/changsongl/master-election/log"
	"testing"
)

func TestLogger(t *testing.T) {
	l := log.newLogger(log.LevelInfo, nil)
	l.Debug("this a debug msg")
	l.Info("this a info msg")
	l.Error("this a error msg")
}
