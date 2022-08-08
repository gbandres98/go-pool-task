package pool

import (
	"context"
	"sync/atomic"
)

type worker struct {
	id                int
	ctx               context.Context
	jobChan           <-chan job
	activeWorkerCount *atomic.Int32
}

func (w *worker) run() {
	go func() {
		for {
			select {
			case job := <-w.jobChan:
				w.activeWorkerCount.Add(1)
				r, err := job.task()
				if err != nil {
					job.errChan <- err
					job.errorCount.Add(1)
				} else {
					job.resChan <- r
					job.successCount.Add(1)
				}
				w.activeWorkerCount.Add(-1)
			case <-w.ctx.Done():
				return
			}
		}
	}()
}
