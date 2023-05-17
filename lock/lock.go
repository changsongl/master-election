package lock

type MasterLock interface {
	Lock(info *Info) (isSuccess bool, err error)

	UnLock(info *Info) (isSuccess bool, err error)

	WriteHeartbeat(info *Info) error

	CurrentMaster() (*Info, error)
}
