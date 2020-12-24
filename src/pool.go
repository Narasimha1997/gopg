package main

import (
	"log"
	"net/http"
	"os"
	"sync"
)

//WorkerPool The work pool executor
type WorkerPool struct {
	queue     *ConcurrentQueue
	nWorkers  int
	queueSize int

	workingGroup *sync.WaitGroup
}

//HandlerFunction handles the http request
type HandlerFunction func(writer *http.ResponseWriter, reader *http.Request, ch chan<- bool)

//WorkType represents the work
type WorkType struct {
	writer  *http.ResponseWriter
	reader  *http.Request
	handler *HandlerFunction
	channel chan<- bool
}

func poolWorker(wg *sync.WaitGroup, queue *ConcurrentQueue, idx int) {
	defer wg.Done()

	log.Printf("Created worker %d\n", idx)

	for {
		//wait for job
		val, err := queue.dequeue()
		if err != nil {
			log.Fatal(err)
			os.Exit(0)
		}

		httpWork := val.(WorkType)

		log.Printf("work processing in progress by worker %d\n", idx)
		//execute the work function
		(*httpWork.handler)(httpWork.writer, httpWork.reader, httpWork.channel)
	}
}

//SubmitJob submits a new job to the work-queue
func (wokerPool *WorkerPool) SubmitJob(
	w *http.ResponseWriter,
	r *http.Request,
	handler *HandlerFunction,
	channel chan<- bool) {
	work := WorkType{}
	work.writer = w
	work.reader = r
	work.handler = handler
	work.channel = channel

	wokerPool.queue.enqueue(work)

}

//NewWorkerPool Creates a new workerpool and returns the struct
func NewWorkerPool(nWorkers int, queueSize int) *WorkerPool {
	var wg sync.WaitGroup

	queue := NewConcurrentQueue(uint32(queueSize))

	for idx := 0; idx < nWorkers; idx++ {
		wg.Add(1)
		go poolWorker(&wg, queue, idx)
	}

	workerPool := WorkerPool{}
	workerPool.queue = queue
	workerPool.workingGroup = &wg
	workerPool.nWorkers = nWorkers
	workerPool.queueSize = queueSize

	return &workerPool
}
