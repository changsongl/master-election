package master

import (
	"time"

	"github.com/changsongl/master-election/log"
)

const (
	defaultHeartBeat = time.Second * 6

	defaultVersion = "default"

	defaultHeartBeatMultiplier = 3
)

type config struct {
	Heartbeat           time.Duration
	HeartbeatMultiplier int
	Version             string

	MasterStartHook func(epoch uint64)
	MasterEndHook   func(epoch uint64)

	DefaultLoggerLogLevel log.Level
	Logger                log.Logger
}

func newDefaultConfig() *config {
	return &config{
		Heartbeat:             defaultHeartBeat,
		Version:               defaultVersion,
		HeartbeatMultiplier:   defaultHeartBeatMultiplier,
		DefaultLoggerLogLevel: log.LevelInfo,
	}
}

type Option interface {
	apply(c *config)
}

type optFunc func(c *config)

func (f optFunc) apply(c *config) {
	f(c)
}

func OptionHeartbeat(h time.Duration) Option {
	return optFunc(func(c *config) {
		c.Heartbeat = h
	})
}

func OptionHeartBeatMultiplier(m int) Option {
	return optFunc(func(c *config) {
		c.HeartbeatMultiplier = m
	})
}

func OptionVersion(version string) Option {
	return optFunc(func(c *config) {
		c.Version = version
	})
}

func OptionMasterStartHook(f func(epoch uint64)) Option {
	return optFunc(func(c *config) {
		c.MasterStartHook = f
	})
}

func OptionMasterEndHook(f func(epoch uint64)) Option {
	return optFunc(func(c *config) {
		c.MasterEndHook = f
	})
}

func OptionLogger(l log.Logger) Option {
	return optFunc(func(c *config) {
		c.Logger = l
	})
}

func OptionDefaultLoggerLogLevel(l log.Level) Option {
	return optFunc(func(c *config) {
		c.DefaultLoggerLogLevel = l
	})
}
