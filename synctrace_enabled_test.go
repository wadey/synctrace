//go:build synctrace
// +build synctrace

package synctrace

import (
	"testing"
)

func TestLockOrderConsistent(t *testing.T) {
	var a = Mutex{Name: "1-a"}
	var b = Mutex{Name: "1-b"}
	var c = Mutex{Name: "1-c"}

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
	ChanDebugRecvLock("1-d")
	ChanDebugRecvUnlock("1-d")
	c.Unlock()
	a.Unlock()

	a.Lock()
	b.Lock()
	c.Lock()
	ChanDebugRecvLock("1-d")
	ChanDebugRecvUnlock("1-d")
	c.Unlock()
	b.Unlock()
	a.Unlock()

	a.Lock()
	ChanDebugSend("1-d")
	a.Unlock()
}

func TestLockOrderReentrant(t *testing.T) {
	var a = Mutex{Name: "2-a"}

	a.Lock()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	a.Lock()
}

func TestLockOrderInconsistent(t *testing.T) {
	var a = Mutex{Name: "3-a"}
	var b = Mutex{Name: "3-b"}
	var c = Mutex{Name: "3-c"}

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

func TestLockOrderInconsistentChan(t *testing.T) {
	var a = Mutex{Name: "4-a"}

	ChanDebugRecvLock("4-b")
	a.Lock()
	a.Unlock()
	ChanDebugRecvUnlock("4-b")

	a.Lock()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	ChanDebugSend("4-b")
}
