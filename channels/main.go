package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	//"os"
	//"strconv"
	"time"
)

var (
	//maxWorkers, _ = strconv.Atoi(os.Getenv("MAX_WORKERS"))
	//maxQueue, _   = strconv.Atoi(os.Getenv("MAX_QUEUE"))
	maxWorkers = 2
	maxQueue   = 5
)

// Job represents the job to be run
type Job struct {
	Command string `json:"command"`
}

// A buffered channel that we can send work requests on.
var JobQueue chan Job

// Worker represents the worker that executes the job
type Worker struct {
	Wid        int
	WorkerPool chan chan Job
	JobChannel chan Job
	quit       chan bool
}

func NewWorker(workerPool chan chan Job) Worker {
	return Worker{
		WorkerPool: workerPool,
		JobChannel: make(chan Job),
		quit:       make(chan bool)}
}

// Start method starts the run loop for the worker, listening for a quit channel in
// case we need to stop it
func (w Worker) Start() {
	log.Print("Worker Start !")
	go func() {
		for {
			// register the current worker into the worker queue.
			w.WorkerPool <- w.JobChannel

			select {
			case job := <-w.JobChannel:
				// we have received a work request.
				log.Printf("Worker %d take the job !", w.Wid)
				log.Printf("%s", job.Command)
				time.Sleep(1000 * time.Millisecond)
				log.Printf("Worker %d wake up", w.Wid)

			case <-w.quit:
				// we have received a signal to stop
				return
			}
		}
	}()
}

// Stop signals the worker to stop listening for work requests.
func (w Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}

type Dispatcher struct {
	// A pool of workers channels that are registered with the dispatcher
	WorkerPool chan chan Job
}

func NewDispatcher() *Dispatcher {
	pool := make(chan chan Job, maxWorkers)
	return &Dispatcher{WorkerPool: pool}
}

func (d *Dispatcher) Run() {
	log.Print("Run dispatcher")
	// starting n number of workers
	for i := 0; i < maxWorkers; i++ {
		log.Printf("Create worker %d", i)
		worker := NewWorker(d.WorkerPool)
		worker.Wid = i
		worker.Start()
	}

	go d.dispatch()
}

func (d *Dispatcher) dispatch() {
	log.Print("Dispatch on")
	for {
		select {
		case job := <-JobQueue:
			log.Print("A Job has been detected")
			// a job request has been received

			// JNU : Start a new thread which will wait a slot in the channel, uncomment to make thread blocking instead of http client
			//go func(job Job) {
			// try to obtain a worker job channel that is available.
			// this will block until a worker is idle
			jobChannel := <-d.WorkerPool

			// dispatch the job to the worker job channel
			jobChannel <- job
			//}(job)
		}
	}
}

type Configuration struct {
	ListenAddress string
	PathPrefix    string
}
type msgError struct {
	Code int   `json:"code"`
	Msg  error `json:"msg"`
}

func main() {
	config := Configuration{ListenAddress: ":8080", PathPrefix: ""}
	r := mux.NewRouter()

	d := NewDispatcher()
	d.Run()
	JobQueue = make(chan Job, maxQueue)

	var EmitStackTask = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var work Job
		encOutput := json.NewEncoder(w)
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&work)
		if err != nil {
			log.Print("Unable to decode")
			w.WriteHeader(http.StatusInternalServerError)
			encOutput.Encode(&msgError{
				Code: 500,
				Msg:  err,
			})
			return
		}
		log.Printf("EnQueue %s", work.Command)
		// Push the work onto the queue.
		JobQueue <- work
		log.Printf("Current job queue size : %d", len(JobQueue))
		return
	})

	r.Handle("/jobs", EmitStackTask).Methods("PUT")

	log.Println("Listening on", config.ListenAddress)
	log.Fatal(http.ListenAndServe(config.ListenAddress, r))

}
