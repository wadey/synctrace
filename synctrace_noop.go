//go:build !synctrace
// +build !synctrace

package synctrace

import "sync"

type Key = string

type RWMutex = sync.RWMutex
type Mutex = sync.Mutex

func NewRWMutex(Key) RWMutex {
	return RWMutex{}
}

func NewMutex(Key) Mutex {
	return Mutex{}
}

func ChanDebugRecvLock(name string)   {}
func ChanDebugRecvUnlock(name string) {}
func ChanDebugSend(name string)       {}
