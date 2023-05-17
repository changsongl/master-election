package mysql

import (
	"github.com/changsongl/master-election/lock"
	"github.com/changsongl/master-election/log"
	"gorm.io/gorm"
	"time"
)

type Config struct {
	BaseDSN string

	User     string
	Password string
	Host     string
	Port     int

	DBName    string
	TableName string

	CreateDB bool

	MaxWait            time.Duration
	MaxRetries         int
	MaxOpenConnections int

	logger log.Logger
}

type MasterLock struct {
	c *Config

	logger log.Logger

	db *gorm.DB
}

func NewMasterLock(c *Config) lock.MasterLock {
	l := &MasterLock{
		c: c,
	}
	
	return l
}

func (m *MasterLock) Lock(info *lock.Info) (isSuccess bool, err error) {
	//TODO implement me
	panic("implement me")
}

func (m *MasterLock) UnLock(info *lock.Info) (isSuccess bool, err error) {
	//TODO implement me
	panic("implement me")
}

func (m *MasterLock) WriteHeartbeat(info *lock.Info) error {
	//TODO implement me
	panic("implement me")
}

func (m *MasterLock) CurrentMaster() (*lock.Info, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MasterLock) SetLogger(l log.Logger) {
	m.logger = l
}
