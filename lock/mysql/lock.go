package mysql

import (
	"errors"
	"fmt"
	"time"

	"github.com/changsongl/master-election/lock"
	"github.com/changsongl/master-election/log"

	sqldriver "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	ErrorUpdateHeartbeatFailed = errors.New("update heartbeat failed")
)

const (
	DefaultDBName             = "master_election"
	DefaultTableName          = "master_lock"
	DefaultMaxOpenConnections = 4
	DefaultMaxWait            = time.Second * 5

	DefaultTimeout      = 10
	DefaultReadTimeout  = 10
	DefaultWriteTimeout = 10
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

	ConnMaxLifeTime    time.Duration
	MaxOpenConnections int

	Timeout      int
	ReadTimeout  int
	WriteTimeout int

	RowID uint
}

type MasterLock struct {
	c *Config

	logger log.Logger

	heartbeat           time.Duration
	heartbeatMultiplier int
	masterID            string

	checker lock.Checker

	db *gorm.DB
}

func (c *Config) init() {
	if c.Port == 0 {
		c.Port = 3306
	}

	if c.ConnMaxLifeTime == 0 {
		c.ConnMaxLifeTime = DefaultMaxWait
	}

	if c.DBName == "" {
		c.DBName = DefaultDBName
	}

	if c.TableName == "" {
		c.TableName = DefaultTableName
	}

	if c.BaseDSN == "" {
		c.BaseDSN = fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/?parseTime=true&timeout=%ds&readTimeout=%ds&writeTimeout=%ds",
			c.User, c.Password, c.Host, c.Port, c.Timeout, c.ReadTimeout, c.WriteTimeout)
	}

	if c.MaxOpenConnections <= 0 {
		c.MaxOpenConnections = DefaultMaxOpenConnections
	}

	if c.Timeout == 0 {
		c.Timeout = DefaultTimeout
	}

	if c.ReadTimeout == 0 {
		c.ReadTimeout = DefaultReadTimeout
	}

	if c.WriteTimeout == 0 {
		c.WriteTimeout = DefaultWriteTimeout
	}
}

func NewMasterLock(c *Config) lock.MasterLock {
	c.init()

	l := &MasterLock{
		c: c,
	}

	return l
}

func (m *MasterLock) Lock(info *lock.Info) (isSuccess bool, err error) {
	curInfo, err := m.CurrentMaster()
	if err != nil && err != lock.ErrorCurrentlyNoMaster {
		m.logger.Errorf("MasterLock.Lock failed: %s", err)
		return false, err
	}

	now := time.Now()

	defer func() {
		if isSuccess {
			info.SetStartAt(now)
		}
	}()

	if err == lock.ErrorCurrentlyNoMaster {
		isSuccess, err = m.insertMaster(info, now)
		if err != nil {
			m.logger.Errorf("MasterLock.Lock m.insertMaster failed: %s", err)
			return false, err
		}

		return isSuccess, nil
	}

	if m.checker.IsValid(curInfo) {
		return false, nil
	}

	isSuccess, err = m.updateMaster(curInfo, info, now)
	if err != nil {
		m.logger.Errorf("MasterLock.Lock m.updateMaster failed: %s", err)
		return false, err
	}

	return isSuccess, nil
}

func (m *MasterLock) insertMaster(info *lock.Info, now time.Time) (success bool, err error) {
	sql := fmt.Sprintf(
		"INSERT INTO %s (id, master_id, version, ip, started_at, last_heartbeat) VALUES (?, ?, ?, ?, ?, ?)",
		m.getTableFullName(),
	)

	err = m.db.Table(m.getTableFullName()).Exec(
		sql, m.getRowID(), info.MasterID, info.Version, info.IP, now, now).Error

	if err != nil {
		if errMysql, ok := err.(*sqldriver.MySQLError); ok && errMysql.Number == 1062 {
			return false, nil
		}

		m.logger.Errorf("MasterLock.insertMaster failed: %s", err)
		return false, err
	}

	return true, nil
}

func (m *MasterLock) updateMaster(curMaster, info *lock.Info, now time.Time) (success bool, err error) {
	result := m.db.Table(m.getTableFullName()).
		Where("id = ? AND master_id = ? AND version = ? AND started_at = ? AND last_heartbeat = ? AND ip = ?",
			m.getRowID(), curMaster.MasterID, curMaster.Version,
			curMaster.StartedAt, curMaster.LastHeartbeat, curMaster.IP).
		Updates(map[string]interface{}{
			"master_id":      info.MasterID,
			"version":        info.Version,
			"started_at":     now,
			"last_heartbeat": now,
			"ip":             info.IP,
		})

	if result.Error != nil {
		m.logger.Errorf("MasterLock.updateMaster failed: %s", err)
		return false, err
	}

	return result.RowsAffected > 0, nil
}

func (m *MasterLock) UnLock(info *lock.Info) (isSuccess bool, err error) {
	result := m.db.Table(m.getTableFullName()).
		Where("id = ? AND master_id = ?", m.getRowID(), info.MasterID).Delete(&lock.Info{})

	if result.Error != nil {
		m.logger.Errorf("MasterLock.UnLock failed: %s", result.Error)
		return false, result.Error
	}

	return result.RowsAffected > 0, nil
}

func (m *MasterLock) WriteHeartbeat(info *lock.Info) error {
	now := time.Now()

	result := m.db.Table(m.getTableFullName()).
		Where("id = ? AND master_id = ?", m.getRowID(), info.MasterID).
		Update("last_heartbeat", now)

	if result.Error != nil {
		m.logger.Errorf("MasterLock.WriteHeartbeat failed: %s", result.Error)
		return result.Error
	} else if result.RowsAffected == 0 {
		m.logger.Errorf("MasterLock.WriteHeartbeat RowsAffected 0: %s", ErrorUpdateHeartbeatFailed)
		return ErrorUpdateHeartbeatFailed
	}

	info.LastHeartbeat = now

	return nil
}

func (m *MasterLock) CurrentMaster() (*lock.Info, error) {
	info := &lock.Info{}
	err := m.db.Table(m.getTableFullName()).Where("id = ?", m.getRowID()).First(info).Error
	if err == gorm.ErrRecordNotFound {
		return nil, lock.ErrorCurrentlyNoMaster
	} else if err != nil {
		return nil, err
	}

	return info, nil
}

func (m *MasterLock) Init(c *lock.MasterLockConfig) error {
	m.logger = c.Log
	m.heartbeat = c.Heartbeat
	m.heartbeatMultiplier = c.HeartbeatMultiplier
	m.masterID = c.MasterID

	m.checker = lock.NewChecker(c.MasterID, c.Heartbeat, c.HeartbeatMultiplier)

	db, err := m.newDatabase()
	if err != nil {
		m.logger.Errorf("MasterLock.Init m.newDatabase failed: %s", err)
		return err
	}

	m.db = db

	return nil
}

func (m *MasterLock) getRowID() uint {
	return m.c.RowID
}

func (m *MasterLock) getTableFullName() string {
	return fmt.Sprintf("%s.%s", m.getDBName(), m.getTableName())
}

func (m *MasterLock) getTableName() string {
	return m.c.TableName
}

func (m *MasterLock) getDBName() string {
	return m.c.DBName
}

func (m *MasterLock) newDatabase() (*gorm.DB, error) {
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN: m.c.BaseDSN,
	}), &gorm.Config{})
	if err != nil {
		m.logger.Errorf("MasterLock.newDatabase gorm.Open failed: %s", err)
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		m.logger.Errorf("MasterLock.newDatabase db.DB failed: %s", err)
		return nil, err
	}

	if m.c.ConnMaxLifeTime != 0 {
		sqlDB.SetConnMaxLifetime(m.c.ConnMaxLifeTime)
	}

	sqlDB.SetMaxOpenConns(m.c.MaxOpenConnections)

	if err = m.createDB(db); err != nil {
		m.logger.Errorf("MasterLock.newDatabase m.createDB failed: %s", err)
		return nil, err
	}

	if err = m.createTable(db); err != nil {
		m.logger.Errorf("MasterLock.newDatabase m.createTable failed: %s", err)
		return nil, err
	}

	return db, nil
}

func (m *MasterLock) createDB(db *gorm.DB) error {
	if !m.c.CreateDB {
		return nil
	}

	query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", m.getDBName())
	if err := db.Exec(query).Error; err != nil {
		m.logger.Errorf("MasterLock.createDB db.Exec failed: %s", err)
		return err
	}

	return nil
}

func (m *MasterLock) createTable(db *gorm.DB) error {

	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (`+
		`id INT NOT NULL PRIMARY KEY,`+
		`master_id CHAR(36) UNIQUE,`+
		`version VARCHAR(255),`+
		`ip VARCHAR(16),`+
		`started_at TIMESTAMP NULL,`+
		`last_heartbeat TIMESTAMP NULL`+
		`);`, m.getTableFullName())

	if err := db.Exec(query).Error; err != nil {
		m.logger.Errorf("MasterLock.createTable db.Exec failed: %s", err)
		return err
	}

	return nil
}
