package pool

import (
	"context"
	"sync/atomic"
)

type Queue interface {
	AddJob(func() (any, error))
	ResultChannel() chan any
	ErrorChannel() chan error
	JobCount() int32
	AssignedJobCount() int32
	QueuedJobCount() int
	SuccessCount() int32
	ErrorCount() int32
	FinishedJobCount() int32
}

type queue struct {
	pool         *pool
	ctx          context.Context
	queueChan    chan job
	resChan      chan any
	errChan      chan error
	jobs         atomic.Int32
	assignedJobs atomic.Int32
	successes    atomic.Int32
	errors       atomic.Int32
}

func (q *queue) run() {
	go func() {
		for {
			select {
			case job := <-q.queueChan:
				q.pool.jobChan <- job
				q.assignedJobs.Add(1)
			case <-q.ctx.Done():
				return
			}
		}
	}()
}

func (q *queue) AddJob(task func() (any, error)) {
	job := job{
		task:         task,
		resChan:      q.resChan,
		errChan:      q.errChan,
		successCount: &q.successes,
		errorCount:   &q.errors,
	}

	q.jobs.Add(1)

	q.queueChan <- job
}

func (q *queue) ResultChannel() chan any {
	return q.resChan
}

func (q *queue) ErrorChannel() chan error {
	return q.errChan
}

func (q *queue) JobCount() int32 {
	return q.jobs.Load()
}

func (q *queue) QueuedJobCount() int {
	return len(q.queueChan)
}

func (q *queue) AssignedJobCount() int32 {
	return q.assignedJobs.Load()
}

func (q *queue) SuccessCount() int32 {
	return q.successes.Load()
}

func (q *queue) ErrorCount() int32 {
	return q.errors.Load()
}

func (q *queue) FinishedJobCount() int32 {
	return q.successes.Load() + q.errors.Load()
}
