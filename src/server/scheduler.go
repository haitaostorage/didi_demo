package main

import "github.com/donnie4w/go-logger/logger"

const MAX_WORKER int = 1

type Scheduler struct {
	WorkPool chan chan Job
}

/*var timeout chan bool = make(chan bool,1)

func TimeOut(){
	time.Sleep(1e9)
	timeout<-true
}*/

func NewScheduler() *Scheduler {
	return &Scheduler{
		WorkPool: make(chan chan Job, MAX_WORKER)}
}

func (this *Scheduler) Run() {
	for i := 0; i < MAX_WORKER; i++ {
		worker := NewWorker(this.WorkPool)
		worker.Start()
	}
	go this.schedule()
}

func (this *Scheduler) schedule() {
	for {
		job := <-JobQueue
		logger.Info("scheduler receive a job from JobQueue", job)
		select {
		case jobchan := <-this.WorkPool:
			logger.Info("scheduler select a worker to process job", job)
			jobchan <- job

		}
	}
}
