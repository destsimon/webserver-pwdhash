package worker

import (
	"testing"
	"sync/atomic"
)

func TestWorker_AddJobsAndStop_WorkerPoolFinishesAllJobsBeforeShuttingDown(t *testing.T) {
	//setup a worker pool
	count := uint64(0)

	wp := NewWorkerPool(1000, 100, func(id uint64) { atomic.AddUint64(&count, 1)})
	go wp.InitWorkerPool()

	numberOfJobs := 1000

	for i := 1; i<=1000; i++ {
		wp.AddJob(uint64(i))
	}

	wp.Stop()

	if count != uint64(numberOfJobs) {
		t.Errorf("unexpected number of worker function call. expected=%d, got=%s", numberOfJobs, count)
	}

}
