//go:build synctrace
// +build synctrace

package synctrace

import (
	"fmt"
	"log/slog"
	"runtime"
	"sync"

	"github.com/heimdalr/dag"
	"github.com/timandy/routine"
)

type Key = dag.IDInterface

type readKey struct {
	id string
	v  mutexValue
}

type mutexValue struct {
	file     string
	line     int
	readOnly bool
}

func (m mutexValue) String() string {
	return fmt.Sprintf("%s:%d", m.file, m.line)
}
func (m mutexValue) ID(key Key) string {
	if m.readOnly {
		return fmt.Sprintf("%s:read:%s", key.ID(), m)
	} else {
		return key.ID()
	}
}

var threadLocal routine.ThreadLocal = routine.NewThreadLocalWithInitial(func() any { return map[Key]mutexValue{} })

var locks = dag.NewDAG()
var readLocks = map[Key][]readKey{}

func NewRWMutex(name string) RWMutex {
	return RWMutex{Name: name}
}

func NewMutex(name string) Mutex {
	return Mutex{Name: name}
}

type RWMutex struct {
	sync.RWMutex
	Name string
	id   string
}

type Mutex struct {
	sync.Mutex
	Name string
	id   string
}

func (m *RWMutex) ID() string {
	if m.id == "" {
		if m.Name != "" {
			m.id = fmt.Sprintf("%s (%p)", m.Name, m)
		} else {
			m.id = fmt.Sprintf("%p", m)
		}
	}
	return m.id
}

func (m *RWMutex) String() string {
	return m.ID()
}

func (m *Mutex) ID() string {
	if m.id == "" {
		if m.Name != "" {
			m.id = fmt.Sprintf("%s (%p)", m.Name, m)
		} else {
			m.id = fmt.Sprintf("%p", m)
		}
	}
	return m.id
}

func (m *Mutex) String() string {
	return m.ID()
}

func (m *readKey) ID() string {
	return fmt.Sprintf("%s:%s", m.id, m.v)
}

func (m *readKey) String() string {
	return m.ID()
}

func alertMutex(err error) {
	panic(err)
}

func checkMutex(state map[Key]mutexValue, add Key, v mutexValue) Key {
	_, err := locks.AddVertex(add)
	if err != nil {
		switch err.(type) {
		case dag.VertexDuplicateError:
			// ignore
		default:
			panic(err)
		}
	}

	aid := add.ID()

	if _, ok := add.(*RWMutex); ok {
		if v.readOnly {
			// use a non-unique key
			radd := readKey{id: add.ID(), v: v}
			readLocks[add] = append(readLocks[add], radd)
			raid := v.ID(add)

			err := locks.AddVertexByID(raid, radd)
			if err != nil {
				switch err.(type) {
				case dag.VertexDuplicateError:
					// ignore
				default:
					panic(err)
				}
			}
			err = locks.AddEdge(aid, raid)
			if err != nil {
				switch err.(type) {
				case dag.EdgeLoopError:
					alertMutex(fmt.Errorf("grabbing read-lock %s but already have these locks: %v. Would cause a DAG loop", raid, state))
				case dag.EdgeDuplicateError:
					// ignore
				default:
					panic(err)
				}
			}
			aid = raid
		}
	}

	for k, v := range state {
		vid := v.ID(k)
		slog.Info("adding", "src", vid, "dst", aid)
		err := locks.AddEdge(vid, aid)
		if err != nil {
			switch err.(type) {
			case dag.SrcDstEqualError:
				alertMutex(fmt.Errorf("reentrant lock of %s, already have these locks: %v", aid, state))
			case dag.EdgeLoopError:
				alertMutex(fmt.Errorf("grabbing lock %s but already have these locks: %v. Would cause a DAG loop", aid, state))
			case dag.EdgeDuplicateError:
				// ignore
			default:
				panic(err)
			}
		}
	}

	return add
}

func newMutexValue() (v mutexValue) {
	_, v.file, v.line, _ = runtime.Caller(2)
	return v
}

func (s *RWMutex) Lock() {
	var key Key = s
	m := threadLocal.Get().(map[Key]mutexValue)
	v := newMutexValue()
	checkMutex(m, key, v)
	m[key] = v
	s.RWMutex.Lock()
}

func (s *RWMutex) Unlock() {
	var key Key = s
	m := threadLocal.Get().(map[Key]mutexValue)
	delete(m, key)
	s.RWMutex.Unlock()
}

func (s *RWMutex) RLock() {
	var key Key = s
	m := threadLocal.Get().(map[Key]mutexValue)
	v := newMutexValue()
	v.readOnly = true
	checkMutex(m, key, v)
	m[key] = v
	s.RWMutex.RLock()
}

func (s *RWMutex) RUnlock() {
	var key Key = s
	m := threadLocal.Get().(map[Key]mutexValue)
	delete(m, key)
	s.RWMutex.RUnlock()
}

func (s *Mutex) Lock() {
	var key Key = s
	m := threadLocal.Get().(map[Key]mutexValue)
	v := newMutexValue()
	checkMutex(m, key, v)
	m[key] = v
	s.Mutex.Lock()
}

func (s *Mutex) Unlock() {
	var key Key = s
	m := threadLocal.Get().(map[Key]mutexValue)
	delete(m, key)
	s.Mutex.Unlock()
}

func ChanDebugRecvStart(key Key) {
	m := threadLocal.Get().(map[Key]mutexValue)
	v := newMutexValue()
	checkMutex(m, key, v)
	m[key] = v
}

func ChanDebugRecvFinished(key Key) {
	m := threadLocal.Get().(map[Key]mutexValue)
	delete(m, key)
}

func ChanDebugSend(key Key) {
	m := threadLocal.Get().(map[Key]mutexValue)
	v := newMutexValue()
	checkMutex(m, key, v)
}
