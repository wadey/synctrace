//go:build synctrace
// +build synctrace

package synctrace

import (
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"sync"

	"github.com/heimdalr/dag"
	"github.com/timandy/routine"
)

type Key = dag.IDInterface

type stringKey string

func (s stringKey) ID() string {
	return string(s)
}

type mutexValue struct {
	file string
	line int
}

func (m mutexValue) String() string {
	return fmt.Sprintf("%s:%d", m.file, m.line)
}

var threadLocal = routine.NewThreadLocalWithInitial(func() map[Key]mutexValue { return map[Key]mutexValue{} })

var locks = dag.NewDAG()

type edgeKey struct {
	src string
	dst string
}

var edgeLocations = map[edgeKey]map[mutexValue]bool{}
var edgeLocationsLock sync.Mutex

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
			m.id = m.Name
			// m.id = fmt.Sprintf("%s (%p)", m.Name, m)
		} else {
			panic("no name")
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
			// m.id = fmt.Sprintf("%s (%p)", m.Name, m)
			m.id = m.Name
		} else {
			panic("no name")
			m.id = fmt.Sprintf("%p", m)
		}
	}
	return m.id
}

func (m *Mutex) String() string {
	return m.ID()
}

func alertMutex(err error, state map[Key]mutexValue, addID string) {
	f, err := os.CreateTemp("", "*.dot")
	if err != nil {
		slog.Error("failed to create dot file", "error", err)
		f = nil
	} else {
		defer f.Close()
		fmt.Fprintln(f, "digraph locks {")
	}

	edgeLocationsLock.Lock()
	defer edgeLocationsLock.Unlock()

	for e, v := range edgeLocations {
		fmt.Fprintf(os.Stderr, "%s -> %s:\n", e.src, e.dst)
		for mv, _ := range v {
			fmt.Fprintf(os.Stderr, "\t%s\n", mv)
		}
		fmt.Fprintln(os.Stderr)

		if f != nil {
			fmt.Fprintf(f, "\"%s\" -> \"%s\";\n", e.src, e.dst)
		}
	}

	if f != nil {
		for k, _ := range state {
			fmt.Fprintf(f, "\"%s\" -> \"%s\" [color=red];\n", k.ID(), addID)
		}
		fmt.Fprintln(f, "}")
		slog.Info("digraph locks file written", "file", f.Name())
	}

	panic(err)
}

func init() {
	m := threadLocal.Get()
	v := mutexValue{}
	checkMutex(m, stringKey("remote-list"), v)
	m[stringKey("remote-list")] = v
	checkMutex(m, stringKey("hostmap"), v)
	m[stringKey("hostmap")] = v
}

func checkMutex(state map[Key]mutexValue, add Key, v mutexValue) Key {
	_, err := locks.AddVertex(add)
	if err != nil {
		switch err.(type) {
		case dag.VertexDuplicateError, dag.IDDuplicateError:
			// ignore
		default:
			panic(err)
		}
	}

	aid := add.ID()

	for k := range state {
		kid := k.ID()
		err := locks.AddEdge(kid, aid)
		if err != nil {
			switch err.(type) {
			case dag.SrcDstEqualError:
				alertMutex(fmt.Errorf("reentrant lock of %s, already have these locks: %v", aid, state), state, aid)
			case dag.EdgeLoopError:
				alertMutex(fmt.Errorf("grabbing lock %s but already have these locks: %v. Would cause a DAG loop", aid, state), state, aid)
			case dag.EdgeDuplicateError:
				// ignore
			default:
				panic(err)
			}
		} else {
			slog.Info("adding", "src", kid, "dst", aid, "v", v)
			fmt.Fprintln(os.Stderr, locks.String())
		}

		edgeLocationsLock.Lock()
		e := edgeLocations[edgeKey{src: kid, dst: aid}]
		if e == nil {
			e = map[mutexValue]bool{}
			edgeLocations[edgeKey{src: kid, dst: aid}] = e
		}
		if !e[v] {
			e[v] = true
			slog.Info("new loc", "src", kid, "dst", aid, "e", edgeLocations)
		}
		edgeLocationsLock.Unlock()
	}

	return add
}

func newMutexValue() (v mutexValue) {
	_, v.file, v.line, _ = runtime.Caller(2)
	return v
}

func (s *RWMutex) Lock() {
	var key Key = s
	m := threadLocal.Get()
	v := newMutexValue()
	checkMutex(m, key, v)
	m[key] = v
	s.RWMutex.Lock()
}

func (s *RWMutex) Unlock() {
	var key Key = s
	m := threadLocal.Get()
	delete(m, key)
	s.RWMutex.Unlock()
}

func (s *RWMutex) RLock() {
	var key Key = s
	m := threadLocal.Get()
	v := newMutexValue()
	checkMutex(m, key, v)
	m[key] = v
	s.RWMutex.RLock()
}

func (s *RWMutex) RUnlock() {
	var key Key = s
	m := threadLocal.Get()
	delete(m, key)
	s.RWMutex.RUnlock()
}

func (s *Mutex) Lock() {
	var key Key = s
	m := threadLocal.Get()
	v := newMutexValue()
	checkMutex(m, key, v)
	m[key] = v
	s.Mutex.Lock()
}

func (s *Mutex) Unlock() {
	var key Key = s
	m := threadLocal.Get()
	delete(m, key)
	s.Mutex.Unlock()
}

func ChanDebugRecvLock(name string) {
	key := stringKey(name)
	m := threadLocal.Get()
	v := newMutexValue()
	checkMutex(m, key, v)
	m[key] = v
}

func ChanDebugRecvUnlock(name string) {
	key := stringKey(name)
	m := threadLocal.Get()
	delete(m, key)
}

func ChanDebugSend(name string) {
	key := stringKey(name)
	m := threadLocal.Get()
	v := newMutexValue()
	checkMutex(m, key, v)
}
