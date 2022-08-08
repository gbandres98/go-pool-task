package pool

import (
	"context"
	"sync/atomic"
)

type Pool interface {
	NewQueue(size int, ctx context.Context) Queue
	WorkerCount() int
	ActiveWorkerCount() int32
}

type pool struct {
	ctx           context.Context
	maxWorkers    int
	jobChan       chan job
	workers       int
	activeWorkers atomic.Int32
}

func InitPool(maxWorkers int, ctx context.Context) Pool {
	p := pool{
		ctx:        ctx,
		maxWorkers: maxWorkers,
		jobChan:    make(chan job),
	}

	for i := 0; i < maxWorkers; i++ {
		w := worker{
			id:                p.workers,
			ctx:               p.ctx,
			jobChan:           p.jobChan,
			activeWorkerCount: &p.activeWorkers,
		}

		w.run()

		p.workers++
	}

	return &p
}

func (p *pool) NewQueue(size int, ctx context.Context) Queue {
	if ctx == nil {
		ctx = p.ctx
	}

	q := &queue{
		pool:      p,
		ctx:       ctx,
		queueChan: make(chan job, size),
		resChan:   make(chan any, size),
		errChan:   make(chan error, size),
	}

	q.run()

	return q
}

func (p *pool) WorkerCount() int {
	return p.workers
}

func (p *pool) ActiveWorkerCount() int32 {
	return p.activeWorkers.Load()
}
