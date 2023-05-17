package lock

import "github.com/changsongl/master-election/log"

type MasterLock interface {
	Lock(info *Info) (isSuccess bool, err error)

	UnLock(info *Info) (isSuccess bool, err error)

	WriteHeartbeat(info *Info) error

	CurrentMaster() (*Info, error)

	SetLogger(l log.Logger)
}
