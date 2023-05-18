package safe

import (
	"sync"
)

type Bool struct {
	sync.RWMutex
	state bool
}

func (b *Bool) SetBool(state bool) {
	b.Lock()
	defer b.Unlock()

	b.state = state
}

func (b *Bool) SetWithCond(prevValue, newValue bool) bool {
	b.Lock()
	defer b.Unlock()

	if b.state != prevValue {
		return false
	}

	b.state = newValue
	return true
}

func (b *Bool) SetTrue() {
	b.SetBool(true)
}

func (b *Bool) SetFalse() {
	b.SetBool(false)
}

func (b *Bool) Value() bool {
	b.RLock()
	defer b.RUnlock()

	return b.state
}

func New() *Bool {
	return &Bool{}
}
