package ticker

import (
	"sync"
	"time"
)

type Ticker interface {
	Loop(func())
	Stop()
}

type ticker struct {
	once            sync.Once
	closeChan       chan struct{}
	closeFinishChan chan struct{}
	ticker          *time.Ticker
	heartbeat       time.Duration
}

func (t *ticker) Loop(f func()) {
	f()

	for {
		select {
		case <-t.ticker.C:
		case <-t.closeChan:
			close(t.closeFinishChan)
			return
		}
	}
}

func (t *ticker) Stop() {
	t.ticker.Stop()

	close(t.closeChan)
	<-t.closeFinishChan
}

func New(d time.Duration) Ticker {
	return &ticker{
		once:            sync.Once{},
		closeChan:       make(chan struct{}),
		closeFinishChan: make(chan struct{}),
		heartbeat:       d,
		ticker:          time.NewTicker(d),
	}
}
