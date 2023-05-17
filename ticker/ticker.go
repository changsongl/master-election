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
	once      sync.Once
	closeChan chan struct{}
	ticker    *time.Ticker
	heartbeat time.Duration
}

func (t *ticker) Loop(f func()) {

	for range t.ticker.C {
		f()
	}

	t.once.Do(func() {
		close(t.closeChan)
	})
}

func (t *ticker) Stop() {
	t.ticker.Stop()
	<-t.closeChan
}

func New(d time.Duration) Ticker {
	return &ticker{
		once:      sync.Once{},
		closeChan: make(chan struct{}),
		heartbeat: d,
		ticker:    time.NewTicker(d),
	}
}
