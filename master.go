package master

import (
	"errors"
	"time"

	"github.com/changsongl/master-election/lock"
	"github.com/changsongl/master-election/net"
	"github.com/changsongl/master-election/safe"
	"github.com/changsongl/master-election/ticker"

	"github.com/google/uuid"
)

var (
	ErrMasterIsStarted     = errors.New("master is started before")
	ErrMasterHasNotStarted = errors.New("master has not started")
)

type Master interface {
	Start() error
	Stop() error
	IsMaster() bool

	ID() string
}

type master struct {
	uuid       string
	version    string
	ip         string
	isMaster   *safe.Bool
	masterInfo *lock.Info
	isStarted  *safe.Bool

	masterStartHook func(epoch uint64)
	masterStopHook  func(epoch uint64)

	heartbeat           time.Duration
	heartbeatMultiplier int
	ticker              ticker.Ticker

	logger Logger

	epoch uint64

	lock lock.MasterLock
}

func (m *master) Start() error {
	if !m.isStarted.SetWithCond(false, true) {
		m.logger.Error("master.Start m.isStarted.SetWithCond: %s", ErrMasterIsStarted)
		return ErrMasterIsStarted
	}

	m.ticker.Loop(func() {
		m.masterInfo.SetCurrentHeartbeatTime(time.Now())

		if !m.isMaster.Value() {
			hasMaster, err := m.hasCurrentMaster()
			if err != nil {
				m.logger.Error("master.Start m.hasCurrentMaster: %s", err)
				return
			}

			if hasMaster {
				m.logger.Debug("master.Start m.hasCurrentMaster: true")
				return
			}

			success, err := m.becomeMaster()
			if err != nil {
				m.logger.Error("master.Start m.becomeMaster: %s", err)
				return
			}

			if success {
				m.logger.Info("master.Start m.becomeMaster: success")

				m.isMaster.SetTrue()
				m.masterInfo.SetStartAt(m.masterInfo.CurrentHeartbeatTime)
				m.nextEpoch()
				m.runMasterStartHook()
			} else {
				m.logger.Debug("master.Start m.becomeMaster: failed")
			}

			return
		}

		err := m.lock.WriteHeartbeat(m.masterInfo)
		if err != nil {
			m.logger.Error("master.Start m.lock.WriteHeartbeat: %s", err)

			m.cleanMasterState()

			return
		}

		m.masterInfo.SetLastHeartBeat(m.masterInfo.CurrentHeartbeatTime)

		m.logger.Debug("master.Start m.lock.WriteHeartbeat: %v", m.masterInfo.LastHeartbeat)
	})
	return nil
}

func (m *master) hasCurrentMaster() (bool, error) {
	curMaster, err := m.lock.CurrentMaster()
	if err != nil {
		m.logger.Error("master.hasCurrentMaster m.lock.CurrentMaster: %s", err)
		return false, err
	}

	if curMaster == nil {
		return false, nil
	}

	return curMaster.IsValid(m.getUUID(), m.heartbeat, 2), nil
}

func (m *master) becomeMaster() (bool, error) {
	success, err := m.lock.Lock(m.masterInfo)
	if err != nil {
		m.logger.Error("master.becomeMaster m.masterInfo.SetCurrentHeartbeatTime: %s", err)

		return false, err
	}

	return success, nil
}

func (m *master) cleanMasterState() {
	m.isMaster.SetFalse()
	m.masterInfo.Clean()

	m.runMasterStopHook()
}

func (m *master) Stop() error {
	m.logger.Info("master.Stop start")

	if !m.isStarted.SetWithCond(true, false) {
		m.logger.Error("master.Stop m.isStarted.SetWithCond: %s", ErrMasterHasNotStarted)
		return ErrMasterHasNotStarted
	}

	m.logger.Debug("master.Stop m.ticker.Stop: before call")
	m.ticker.Stop()
	m.logger.Debug("master.Stop m.ticker.Stop: after call")

	success, err := m.lock.UnLock(m.masterInfo)
	if err != nil {
		m.logger.Error("master.Stop m.lock.UnLock: %s", err)
		return err
	}

	m.logger.Info("master.Stop end: is_master: %t", success)

	return nil
}

func (m *master) IsMaster() bool {
	return m.isMaster.Value()
}

func (m *master) ID() string {
	return m.getUUID()
}

func New(l lock.MasterLock) (Master, error) {
	c := newDefaultConfig()

	id, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	uuidStr := id.String()
	ip := net.GetLocalIP()

	return &master{
		uuid:       uuidStr,
		version:    c.Version,
		ip:         ip,
		isMaster:   safe.New(),
		masterInfo: lock.NewInfo(uuidStr, c.Version, ip),
		isStarted:  safe.New(),

		masterStartHook: c.MasterStartHook,
		masterStopHook:  c.MasterEndHook,

		heartbeat:           c.Heartbeat,
		heartbeatMultiplier: c.HeartbeatMultiplier,
		ticker:              ticker.New(c.Heartbeat),

		lock: l,
		
		logger: newLogger(c.DefaultLoggerLogLevel, c.Logger),

		epoch: 0,
	}, nil
}

func (m *master) getUUID() string {
	return m.uuid
}

func (m *master) isMasterByUUID(id string) bool {
	return id == m.getUUID()
}

func (m *master) nextEpoch() {
	m.epoch++
}

func (m *master) runMasterStartHook() {
	if m.masterStartHook == nil {
		return
	}

	go m.masterStartHook(m.epoch)
}

func (m *master) runMasterStopHook() {
	if m.masterStopHook == nil {
		return
	}

	go m.masterStopHook(m.epoch)
}
