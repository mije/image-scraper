package worker

import (
	"runtime"
	"sync"
)

type Task struct {
	Info interface{}
	Fn   func() error
}

type PoolOptions func(p *Pool)


func WithSize(n int) PoolOptions {
	return func(p *Pool) {
		p.size = n
	}
}

func WithConcurrency(n int) PoolOptions {
	return func(p *Pool) {
		p.concurrency = n
	}
}

func WithErrorHandler(h func(interface{}, error)) PoolOptions {
	return func(p *Pool) {
		p.errorHandler = h
	}
}

type Pool struct {
	size         int
	concurrency  int
	wg           sync.WaitGroup
	tasks        chan Task
	errorHandler func(interface{}, error)
}

func NewPool(opts ...PoolOptions) *Pool {
	p := &Pool{
		concurrency: runtime.NumCPU(),
		size: 1<<10,
	}

	for _, fn := range opts {
		fn(p)
	}

	p.tasks = make(chan Task, p.size)

	return p
}

func (p *Pool) Start() {
	for i := 0; i < p.concurrency; i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			for t := range p.tasks {
				err := t.Fn()
				if err != nil && p.errorHandler != nil {
					p.errorHandler(t.Info, err)
				}
			}
		}()
	}
}

func (p *Pool) Stop() {
	if p.tasks != nil {
		close(p.tasks)
		p.wg.Wait()
	}
}

func (p *Pool) Submit(t Task) {
	p.tasks <- t
}