package lock

import "time"

type checker struct {
	heartbeat           time.Duration
	heartbeatMultiplier int

	masterID string
}

type Checker interface {
	IsValid(i *Info) bool
}

func NewChecker(masterID string, heartbeat time.Duration, heartbeatMultiplier int) Checker {
	return &checker{
		heartbeat:           heartbeat,
		heartbeatMultiplier: heartbeatMultiplier,
		masterID:            masterID,
	}
}

func (c *checker) IsValid(i *Info) bool {
	return time.Since(i.LastHeartbeat) < c.heartbeat*time.Duration(c.heartbeatMultiplier) && i.MasterID != c.masterID
}
