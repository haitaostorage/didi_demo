package main

import (
	"github.com/donnie4w/go-logger/logger"
	//"didi_api/models"
	"pool"
)

type Worker struct {
	WorkChan chan chan Job
	JobChan  chan Job
	QuitChan chan bool
}

func NewWorker(pool chan chan Job) *Worker {
	return &Worker{
		WorkChan: pool,
		JobChan:  make(chan Job),
		QuitChan: make(chan bool)}
}

func (this *Worker) Start() {
	go func() {
		for {
			this.WorkChan <- this.JobChan
			//Logger.Println("worker register self JobChannel",this.JobChan)
			select {
			case job := <-this.JobChan:
				logger.Info("worker receive a Job and process", job)
				process(job.PayLoad)
			case <-this.QuitChan:
				return
			}
		}
	}()
}

func (this *Worker) Stop() {
	go func() {
		this.QuitChan <- true
	}()
}

func process(job interface{}) {
	switch job.(type) {
	case *pool.DriverElement:
		logger.Info("worker process job to join driver into pool", job)
		job.(*pool.DriverElement).Process()
	case *pool.OrderElement:
		logger.Info("worker process job to join order into pool", job)
		job.(*pool.OrderElement).Process()
	case *pool.PassengerState:
		logger.Info("worker process job to update passenger state", job)
		job.(*pool.PassengerState).Process()
	case *pool.PoolElement:
		logger.Info("worker process job to release or judge exist", job)
		job.(*pool.PoolElement).Process()
	default:
		logger.Error("job type")
	}

}
