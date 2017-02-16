package pool

import (
	"didi_api/models"
	"encoding/json"
	"github.com/donnie4w/go-logger/logger"
	"github.com/gorilla/websocket"
)

type PoolElement struct {
	utype string
	op    string
	uid   string
}

var (
	DriverChan = make(chan bool)
	OrderChan  = make(chan bool)
)

func NewPoolElement(utype, op, uid string) *PoolElement {
	return &PoolElement{utype, op, uid}
}

func (this *PoolElement) Process() {
	if this.op == "release" {
		this.Release()
	} else {
		this.IsExist()
	}
}

func (this *PoolElement) Release() {
	if this.utype == "driver" {
		this.ReleaseDriver()
	} else {
		this.ReleasePassenger()
	}
}

func (this *PoolElement) ReleaseDriver() {
	elem := Apool.DelArrangedDriverById(this.uid)
	elem.Reset()
	Dpool.AddDriver(elem)
	Sched() //再调度一次
}

func FreshToDB() {

}

func (this *PoolElement) ReleasePassenger() {
	FreshToDB()
	elem := Opool.OrderList[this.uid]
	//ResetOrder(elem)
	elem.Self_x_scale = -1
	elem.Self_y_scale = -1
	elem.Status = models.COMPLETE
	str, _ := json.Marshal(elem)
	elem.Ws.WriteMessage(websocket.TextMessage, str)
	Leave(this.uid, "passenger")
}

func (this *PoolElement) IsExist() {
	if this.utype == "driver" {
		DriverChan <- this.IsDriverExist()
	} else {
		OrderChan <- this.IsOrderExist()
	}

}

func (this *PoolElement) IsDriverExist() bool {
	_, ok := Dpool.DriverList[this.uid]
	if ok {
		return true
	}
	_, ok = Apool.ArrangedList[this.uid]
	if ok {
		return true
	}
	return false
}

func (this *PoolElement) IsOrderExist() bool {
	v, ok := Opool.OrderList[this.uid]
	if ok {
		if v.Status == models.COMPLETE {
			return false
		}
		return true
	}
	return false
}

func Leave(uid string, utype string) {
	if utype == "driver" {
		logger.Info("driver leave pool")
		elem := Dpool.DelDriverById(uid)
		if elem != nil {
			elem.Ws.Close()
			return
		}
		elem = Apool.DelArrangedDriverById(uid)
		if elem != nil {
			elem.Ws.Close()
			return
		}
	} else if utype == "passenger" {
		logger.Info("passenger leave pool")
		elem := Opool.DelOrderById(uid)
		if elem != nil {
			elem.Ws.Close()
			return
		}
	} else {
		return
	}
}
