//go:build !mutex_debug
// +build !mutex_debug

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

func ChanDebugRecv(key Key) {}
func ChanDebugSend(key Key) {}
