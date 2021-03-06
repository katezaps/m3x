// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// Package sync implements synchronization facililites such as worker pools.
package sync

import (
	"time"
)

// Work is a unit of item to be worked on.
type Work func()

// WorkerPool provides a pool for goroutines.
type WorkerPool interface {
	// Init initializes the pool.
	Init()

	// Go waits until the next worker becomes available and executes it.
	Go(work Work)

	// GoIfAvailable performs the work inside a worker if one is available and
	// returns true, or false otherwise.
	GoIfAvailable(work Work) bool

	// GoWithTimeout waits up to the given timeout for a worker to become
	// available, returning true if a worker becomes available, or false
	// otherwise
	GoWithTimeout(work Work, timeout time.Duration) bool
}

type workerPool struct {
	workCh chan struct{}
}

// NewWorkerPool creates a new worker pool.
func NewWorkerPool(size int) WorkerPool {
	return &workerPool{workCh: make(chan struct{}, size)}
}

func (p *workerPool) Init() {
	for i := 0; i < cap(p.workCh); i++ {
		p.workCh <- struct{}{}
	}
}

func (p *workerPool) Go(work Work) {
	token := <-p.workCh
	go func() {
		work()
		p.workCh <- token
	}()
}

func (p *workerPool) GoIfAvailable(work Work) bool {
	select {
	case token := <-p.workCh:
		go func() {
			work()
			p.workCh <- token
		}()
		return true
	default:
		return false
	}
}

func (p *workerPool) GoWithTimeout(work Work, timeout time.Duration) bool {
	select {
	case token := <-p.workCh:
		go func() {
			work()
			p.workCh <- token
		}()
		return true
	case <-time.After(timeout):
		return false
	}
}
