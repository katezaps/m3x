package xsync

import (
	"bytes"
	"fmt"
	"runtime"
	"sync"
	"time"
)

const _StackDepth = 16

// DebugMutex is a RWMutex that tracks its ownership.
type DebugMutex struct {
	m sync.RWMutex
}

// RLock locks DebugMutex for reading.
func (m *DebugMutex) RLock() { m.m.RLock() }

// RUnlock undoes a single RLock call.
func (m *DebugMutex) RUnlock() { m.m.RUnlock() }

// Lock locks DebugMutex for writing.
func (m *DebugMutex) Lock() {
	m.m.Lock()
	insert(m)
}

// Unlock unlocks DebugMutex for writing.
func (m *DebugMutex) Unlock() {
	remove(m)
	m.m.Unlock()
}

// RLocker returns a Locker interface implemented via calls to RLock
// and RUnlock.
func (m *DebugMutex) RLocker() sync.Locker {
	return (*rlocker)(m)
}

type rlocker DebugMutex

func (r *rlocker) Lock()   { r.m.RLock() }
func (r *rlocker) Unlock() { r.m.RUnlock() }

var mutexDebuggingFlag bool

// DisableMutexDebugging turns mutex debugging off.
func DisableMutexDebugging() {
	mutexDebuggingFlag = false
}

// EnableMutexDebugging turns mutex debugging on.
func EnableMutexDebugging() {
	mutexDebuggingFlag = true
}

type lockInfo struct {
	ts time.Time
	cs []uintptr
}

var locks struct {
	sync.Mutex
	m map[*DebugMutex]lockInfo
}

func insert(m *DebugMutex) {
	if !mutexDebuggingFlag {
		return
	}

	r := make([]uintptr, _StackDepth)
	n := runtime.Callers(3, r)

	locks.Lock()
	locks.m[m] = lockInfo{time.Now(), r[:n]}
	locks.Unlock()
}

func remove(m *DebugMutex) {
	if !mutexDebuggingFlag {
		return
	}

	locks.Lock()
	delete(locks.m, m)
	locks.Unlock()
}

func traceback(l []uintptr) string {
	var (
		b    = new(bytes.Buffer)
		n    runtime.Frame
		more = len(l) != 0
	)

	for f := runtime.CallersFrames(l); more; {
		n, more = f.Next()
		fmt.Fprintf(b, "%s:%d\n\t%s\n", n.File, n.Line, n.Function)
	}

	return b.String()
}

// DumpLocks returns all currently locked mutexes.
func DumpLocks() []string {
	var r []string

	locks.Lock()

	for m, l := range locks.m {
		r = append(r, fmt.Sprintf(
			"%p @ %s\n%s", m, time.Since(l.ts), traceback(l.cs)))
	}

	locks.Unlock()

	return r
}

func init() {
	locks.m = make(map[*DebugMutex]lockInfo)
}