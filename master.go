package master

import (
	"errors"
	"time"

	"github.com/changsongl/master-election/lock"
	"github.com/changsongl/master-election/log"
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

	heartbeat time.Duration
	ticker    ticker.Ticker

	logger log.Logger

	epoch uint64

	lock lock.MasterLock
}

func (m *master) Start() error {
	if !m.isStarted.SetWithCond(false, true) {
		m.logger.Errorf("master.Start m.isStarted.SetWithCond: %s", ErrMasterIsStarted)
		return ErrMasterIsStarted
	}

	m.ticker.Loop(func() {

		if !m.isMaster.Value() {
			success, err := m.becomeMaster()
			if err != nil {
				m.logger.Errorf("master.Start m.becomeMaster: %s", err)
				return
			}

			if success {
				m.logger.Infof("master.Start m.becomeMaster: success")

				m.isMaster.SetTrue()
				m.nextEpoch()
				m.runMasterStartHook()
			} else {
				m.logger.Debugf("master.Start m.becomeMaster: failed")
			}

			return
		}

		err := m.lock.WriteHeartbeat(m.masterInfo)
		if err != nil {
			m.logger.Errorf("master.Start m.lock.WriteHeartbeat: %s", err)

			m.cleanMasterState()

			return
		}

		m.logger.Debugf("master.Start m.lock.WriteHeartbeat: %v", m.masterInfo.LastHeartbeat)
	})
	return nil
}

func (m *master) becomeMaster() (bool, error) {
	success, err := m.lock.Lock(m.masterInfo)
	if err != nil {
		m.logger.Errorf("master.becomeMaster m.masterInfo.SetCurrentHeartbeatTime: %s", err)

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
	m.logger.Infof("master.Stop start")

	if !m.isStarted.SetWithCond(true, false) {
		m.logger.Errorf("master.Stop m.isStarted.SetWithCond: %s", ErrMasterHasNotStarted)
		return ErrMasterHasNotStarted
	}

	m.logger.Debugf("master.Stop m.ticker.Stop: before call")
	m.ticker.Stop()
	m.logger.Debugf("master.Stop m.ticker.Stop: after call")

	success, err := m.lock.UnLock(m.masterInfo)
	if err != nil {
		m.logger.Errorf("master.Stop m.lock.UnLock: %s", err)
		return err
	}

	m.logger.Infof("master.Stop end: success: %t", success)

	m.cleanMasterState()

	return nil
}

func (m *master) IsMaster() bool {
	return m.isMaster.Value()
}

func (m *master) ID() string {
	return m.getUUID()
}

func New(l lock.MasterLock, opts ...Option) (Master, error) {
	c := newDefaultConfig()

	for _, opt := range opts {
		opt.apply(c)
	}

	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	uuidStr := id.String()
	ip := net.GetLocalIP()

	logger := log.New(c.DefaultLoggerLogLevel, c.Logger)
	if err = l.Init(&lock.MasterLockConfig{
		Heartbeat:           c.Heartbeat,
		HeartbeatMultiplier: c.HeartbeatMultiplier,
		Log:                 logger,
	}); err != nil {

		logger.Errorf("master.New l.Init failed: %s", err)
		return nil, err
	}

	return &master{
		uuid:       uuidStr,
		version:    c.Version,
		ip:         ip,
		isMaster:   safe.New(),
		masterInfo: lock.NewInfo(uuidStr, c.Version, ip),
		isStarted:  safe.New(),

		masterStartHook: c.MasterStartHook,
		masterStopHook:  c.MasterEndHook,

		heartbeat: c.Heartbeat,
		ticker:    ticker.New(c.Heartbeat),

		lock: l,

		logger: logger,

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
