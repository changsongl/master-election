package lock

import (
	"errors"
	"time"

	"github.com/changsongl/master-election/log"
)

var ErrorCurrentlyNoMaster = errors.New("currently no master")

type MasterLockConfig struct {
	Log log.Logger

	Heartbeat           time.Duration
	HeartbeatMultiplier int

	MasterID string
}

type MasterLock interface {
	Lock(info *Info) (isSuccess bool, err error)

	UnLock(info *Info) (isSuccess bool, err error)

	WriteHeartbeat(info *Info) error

	CurrentMaster() (*Info, error)

	Init(c *MasterLockConfig) error
}
