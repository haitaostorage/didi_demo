package controllers

import "didi_api/models"

type Worker struct{
	WorkChan	chan chan Job
	JobChan		chan Job
	QuitChan 	chan bool
}

func NewWorker(pool chan chan Job) *Worker{
	return &Worker{
		WorkChan: pool,
		JobChan: make(chan Job),
		QuitChan: make(chan bool)}
}



func(this *Worker)Start(){
	go func(){
		for{
			this.WorkChan<-this.JobChan
			//Logger.Println("worker register self JobChannel",this.JobChan)
			select{
				case job := <-this.JobChan:
					Logger.Println("worker receive a Job and process",job)
					process(job.PayLoad)
				case <-this.QuitChan:
					return
			}
		}
	}()
}

func(this *Worker)Stop(){
	go func(){
		this.QuitChan<-true
	}()
}

func process(job interface{}){
	switch job.(type){
		case *DriverElement:
			Logger.Println("worker process job to join driver into pool",job)
			job.(*DriverElement).JoinDriverPool()
			Sched()
		case *OrderElement:
			Logger.Println("worker process job to join order into pool",job)
			job.(*OrderElement).JoinOrderPool()
			job.(*OrderElement).ArrangeDriver()
		case *PassengerState:
			Logger.Println("worker process job to update passenger state",job)
			job.(*PassengerState).UpdatePassengerState()
		case *PoolElement:
			Logger.Println("worker process job to release or judge exist",job)
			job.(*PoolElement).Operate()
		default:
			Logger.Println("job type")
	}

}

func Sched(){
	for _,v := range OrderList{
		if v.Status == models.UNDISPATCH{
			Logger.Println("arrange driver for",v.Puid)
			v.ArrangeDriver()
		}
	}
}


