package controllers
import (
	"didi_api/models"
	"encoding/json"
	"github.com/gorilla/websocket"
       )

type PassengerState struct{
	uid string
	x int
	y int
	status int
}

type PoolElement struct{
	utype 	string
	op	string
	uid	string
}

var (
	DriverChan = make(chan bool)
	OrderChan = make(chan bool)


    )
func (this *PassengerState)UpdatePassengerState(){
	Olock.Lock()
	v,ok := OrderList[this.uid]
	if ok{
		v.Self_x_scale = this.x
		v.Self_y_scale = this.y
		v.Status = this.status
	}
	Olock.Unlock()
}

func (this *PoolElement)Operate(){
	if this.op == "release"{
		this.Release()
	}else{
		this.IsExist()
	}
}

func (this *PoolElement) Release(){
	if this.utype == "driver"{
		this.ReleaseDriver()
	}else{
		this.ReleasePassenger()
	}
}

func (this *PoolElement)ReleaseDriver(){
	elem := ArrangedDriver[this.uid]
	Alock.Lock()
	delete(ArrangedDriver,this.uid)
	Alock.Unlock()
	elem.ResetDriver()
	Dlock.Lock()
	DriverList[this.uid] = elem
	Dlock.Unlock()
	Sched()   //再调度一次
}

func FreshToDB(){

}

func (this *PoolElement)ReleasePassenger(){
	FreshToDB()
	elem := OrderList[this.uid]
	//ResetOrder(elem)
	elem.Self_x_scale = -1
	elem.Self_y_scale = -1
	elem.Status = models.COMPLETE
	str,_ := json.Marshal(elem)
	elem.Ws.WriteMessage(websocket.TextMessage, str)
	Leave(this.uid,"passenger")
}

func (this *PoolElement)IsExist(){
	if this.utype == "driver"{
		DriverChan <-this.IsDriverExist()
	}else{
		OrderChan <-this.IsOrderExist()
	}

}

func (this *PoolElement)IsDriverExist()bool{
	_,ok := DriverList[this.uid]
	if ok{
		return true
	}
	_,ok = ArrangedDriver[this.uid]
	if ok{
		return true
	}
	return false
}

func (this *PoolElement)IsOrderExist()bool{
	Olock.Lock()
	v,ok := OrderList[this.uid]
	Olock.Unlock()
	if ok{
		if v.Status == models.COMPLETE{
			return false
		}
		return true
	}
	return  false
}

