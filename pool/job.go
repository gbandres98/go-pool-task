package pool

import "sync/atomic"

type job struct {
	task         func() (any, error)
	resChan      chan<- any
	errChan      chan<- error
	successCount *atomic.Int32
	errorCount   *atomic.Int32
}
