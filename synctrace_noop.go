//go:build !synctrace
// +build !synctrace

package synctrace

import "sync"

type Key = string

// type syncRWMutex = sync.RWMutex
// type syncMutex = sync.Mutex

func NewRWMutex(Key) sync.RWMutex {
	return sync.RWMutex{}
}

func NewMutex(Key) sync.Mutex {
	return sync.Mutex{}
}

func ChanDebugRecvLock(name string)   {}
func ChanDebugRecvUnlock(name string) {}
func ChanDebugSend(name string)       {}
