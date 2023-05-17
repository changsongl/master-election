package lock

import "time"

type Info struct {
	MasterID string
	Version  string
	IP       string

	CurrentHeartbeatTime time.Time

	StartedAt     time.Time
	LastHeartbeat time.Time
}

func NewInfo(id, version, ip string) *Info {
	return &Info{
		MasterID: id,
		Version:  version,
		IP:       ip,
	}
}

func (i *Info) SetStartAt(t time.Time) *Info {
	i.StartedAt = t
	i.LastHeartbeat = t
	return i
}

func (i *Info) SetLastHeartBeat(t time.Time) *Info {
	i.LastHeartbeat = t
	return i
}

func (i *Info) SetCurrentHeartbeatTime(t time.Time) *Info {
	i.CurrentHeartbeatTime = t
	return i
}

func (i *Info) Clean() *Info {
	i.CurrentHeartbeatTime = time.Unix(0, 0)
	i.StartedAt = time.Unix(0, 0)
	i.LastHeartbeat = time.Unix(0, 0)
	return i
}

func (i *Info) IsValid(masterID string, heartbeat time.Duration, multi int) bool {
	return time.Since(i.LastHeartbeat) < heartbeat*time.Duration(multi) && i.MasterID != masterID
}
