//go:build mutex_debug
// +build mutex_debug

package synctrace

import (
	"testing"
)

func TestLockOrderConsistent(t *testing.T) {
	var a = Mutex{Name: "a"}
	var b = Mutex{Name: "b"}
	var c = Mutex{Name: "c"}

	a.Lock()
	b.Lock()
	b.Unlock()
	a.Unlock()

	a.Lock()
	b.Lock()
	c.Lock()
	c.Unlock()
	b.Unlock()
	a.Unlock()

	b.Lock()
	c.Lock()
	c.Unlock()
	b.Unlock()

	a.Lock()
	c.Lock()
	c.Unlock()
	a.Unlock()
}

func TestRLockOrder(t *testing.T) {
	var a = RWMutex{Name: "a"}
	var b = RWMutex{Name: "b"}

	a.RLock()
	b.RLock()
	b.RUnlock()
	a.RUnlock()

	b.RLock()
	a.RLock()
	a.RLock()
	b.RUnlock()
}

func TestRLockOrderBad(t *testing.T) {
	var a = RWMutex{Name: "a"}
	var b = RWMutex{Name: "b"}

	a.RLock()
	b.RLock()
	b.RUnlock()
	a.RUnlock()

	b.RLock()
	a.RLock()
	a.RUnlock()
	b.RUnlock()

	a.Lock()
	b.RLock()
	b.RUnlock()
	a.Unlock()

	b.Lock()
	a.RLock()
	a.RUnlock()
	b.Unlock()
}

func TestLockOrderReentrant(t *testing.T) {
	var a = Mutex{Name: "a"}

	a.Lock()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	a.Lock()
}

func TestLockOrderInconsistent(t *testing.T) {
	var a = Mutex{Name: "a"}
	var b = Mutex{Name: "b"}
	var c = Mutex{Name: "c"}

	a.Lock()
	b.Lock()
	b.Unlock()
	a.Unlock()

	a.Lock()
	b.Lock()
	c.Lock()
	c.Unlock()
	b.Unlock()
	a.Unlock()

	b.Lock()
	c.Lock()
	c.Unlock()
	b.Unlock()

	a.Lock()
	c.Lock()
	c.Unlock()
	a.Unlock()

	c.Lock()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	a.Lock()
}
