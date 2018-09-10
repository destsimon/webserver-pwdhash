package worker

import (
	"log"
	"sync"
	"time"
)

type WorkerPool struct {
	jobChannel  chan uint64
	workersDone chan bool
	numWorkers  int
	jobFunc     WorkerHandlerFunc
}

type WorkerHandlerFunc func(job uint64)

func NewWorkerPool(numWorkers int, buffSize int, jobFunc WorkerHandlerFunc) *WorkerPool {
	wp := &WorkerPool{
		jobChannel:  make(chan uint64, buffSize),
		workersDone: make(chan bool),
		numWorkers:  numWorkers,
		jobFunc:     jobFunc,
	}

	return wp
}

func (wp *WorkerPool) InitWorkerPool() {
	var wg sync.WaitGroup
	for i := 0; i < wp.numWorkers; i++ {
		wg.Add(1)
		go wp.worker(&wg)
	}
	log.Println("all workers created ... listening for jobs")
	wg.Wait()
	log.Println("all workers are done - send Done signal")
	wp.workersDone <- true
	log.Println("Done signal sent")
}

func (wp *WorkerPool) worker(wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range wp.jobChannel {
		//hard coded sleep time
		time.Sleep(time.Second * 5)
		wp.jobFunc(job)
	}
}

func (wp *WorkerPool) AddJob(job uint64) {
	wp.jobChannel <- job
}

func (wp *WorkerPool) Stop() {
	close(wp.jobChannel)
	<-wp.workersDone
}
