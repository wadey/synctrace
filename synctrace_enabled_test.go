//go:build synctrace
// +build synctrace

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
