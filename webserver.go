package main

import (
	"github.com/destsimon/pwdhash/pwd"
	"github.com/destsimon/pwdhash/worker"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"log"
	"os"
)

var (
	numWorkers = getenvInt(os.Getenv("NUM_WORKERS"), 100)
	bufferSize  = getenvInt(os.Getenv("BUFFER_SIZE"), 1000)
	maxPwdLength = getenvInt(os.Getenv("MAX_PWD_LENGTH"), 64)
	httpPort = getenvInt(os.Getenv("HTTP_PORT"), 8080)
)

func getenvInt(env string, defaultVal int) int {
	if env == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(env)
	if err != nil {
		log.Printf("unable to read environment variable - exit")
		panic(err)
	}
	return v
}

type server struct {
	mux  *http.ServeMux
	stop chan bool

	counter              uint64
	processingTimeMicros uint64
	wp                   *worker.WorkerPool
	pwdStore             *pwd.Store
}

func newServer() *server {
	s := &server{
		mux:      http.NewServeMux(),
		stop:     make(chan bool),
		counter:  0,
		pwdStore: pwd.NewStore(maxPwdLength),
	}

	//register the handler functions
	hashPasswordHandler := http.HandlerFunc(s.hashPasswordHandlerFunc)
	s.mux.Handle("/hash", logRequest(s.trackDuration(hashPasswordHandler)))

	getPasswordHandler := http.HandlerFunc(s.getPasswordHandlerFunc)
	s.mux.Handle("/hash/", logRequest(getPasswordHandler))

	getStatisticsHandler := http.HandlerFunc(s.getStatisticsHandlerFunc)
	s.mux.Handle("/stats", logRequest(getStatisticsHandler))

	shutdownHandler := http.HandlerFunc(s.shutdownHandlerFunc)
	s.mux.Handle("/shutdown", logRequest(shutdownHandler))

	//setup the worker pool
	s.wp = worker.NewWorkerPool(numWorkers, bufferSize, s.hashPasswordWorkerFunc)

	go s.wp.InitWorkerPool()

	return s
}

func main() {
	s := newServer()
	h := &http.Server{Addr: fmt.Sprintf(":%d", httpPort), Handler: s.mux}
	log.Printf("HTTP_PORT=%d, NUM_WORKERS=%d, BUFFER_SIZE=%d, MAX_PWD_LENGTH=%d", httpPort, numWorkers, bufferSize, maxPwdLength)
    log.Println("staring up ...")

	go func() {
		err := h.ListenAndServe()
		if err != nil {
			log.Printf("%s\n", err)
		}
	}()

	<-s.stop
	log.Printf("shutting down ...\n\n")

	h.Shutdown(context.Background())

	//signal the workerPool to stop - blocks until all workers are stopped
	s.wp.Stop()

	log.Printf("shut down complete\n\n")
}

//Decorator, log requests
func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
        log.Printf("processed request method=%s, uri=%s", r.Method, r.RequestURI)
	})
}

//Decorator, tracks the request duration in microseconds
func (s *server) trackDuration(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		atomic.AddUint64(&s.processingTimeMicros, uint64(duration/time.Microsecond))
	})
}

//the function called by the worker to hash a password
func (s *server) hashPasswordWorkerFunc(id uint64) {
	password := s.pwdStore.Get(id)
	password.Hashed = pwd.Hash(password.Clear)
	log.Printf("hashed password=%s, id=%d\n", password.Hashed, id)
	s.pwdStore.Put(id, password)
}

//returns an empty string if the password is not hashed yet, or if the id does not exist
func (s *server) getPasswordHandlerFunc(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/hash/")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	password := s.pwdStore.Get(id)
	if password.Hashed == "" {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, password.Hashed)
	}
}

//adds the password to the worker pool and returns immediately the id
func (s *server) hashPasswordHandlerFunc(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	if r.Form["password"] == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	//remove leading and trailing spaces
	clearPwd := strings.TrimSpace(r.Form["password"][0])

	if err := s.pwdStore.Validate(clearPwd); err != nil {
		log.Printf("rejecting bad password=%s", clearPwd)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id := atomic.AddUint64(&s.counter, 1)

	//create a new password and add it to the store
	password := pwd.Password{Clear: clearPwd}
	s.pwdStore.Put(id, password)

	//create a new password hash job
	s.wp.AddJob(id)

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, strconv.FormatUint(id, 10))
}

type stats struct {
	Total   uint64 `json:"total"`
	Average uint64 `json:"average"`
}

//returns the stats
func (s *server) getStatisticsHandlerFunc(w http.ResponseWriter, r *http.Request) {
	count := atomic.LoadUint64(&s.counter)
	procTime := atomic.LoadUint64(&s.processingTimeMicros)

	stat := &stats{
		Total: count,
	}
	if count > 0 {
		stat.Average = procTime / count
	}
	b, err := json.Marshal(stat)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, string(b))
}

//signals the server to gracefully shut down
func (s *server) shutdownHandlerFunc(w http.ResponseWriter, r *http.Request) {
	s.stop <- true
	w.WriteHeader(http.StatusNoContent)
}
