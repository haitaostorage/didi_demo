package controllers

import(
	"github.com/gorilla/websocket"
	"didi_api/models"
)

type DriverElement struct{
	Ws *websocket.Conn
	Uid string
	Puid string    //该司机接送的乘客id
	Self_x_scale int
	Self_y_scale int
	Status	int
	P_x_scale int
	P_y_scale int
	D_x_scale int
	D_y_scale int
}

func NewDriverElement(ws *websocket.Conn,uid,pid string,x,y,status,p_x,p_y,d_x,d_y int)*DriverElement{
	e := DriverElement{ws,uid,pid,x,y,status,p_x,p_y,d_x,d_y}
	return &e
}

func (this *DriverElement)JoinDriverPool(){
	Dlock.Lock()
	DriverList[this.Uid] = this
	Dlock.Unlock()
}

func (this *DriverElement)UpdateDriverOrderInfo(orderinfo *OrderElement){
	this.Puid = orderinfo.Puid
	this.Status = models.PREPARE
	this.P_x_scale = orderinfo.P_x_scale
	this.P_y_scale = orderinfo.P_y_scale
	this.D_x_scale = orderinfo.D_x_scale
	this.D_y_scale = orderinfo.D_y_scale
}

func (this *DriverElement)ResetDriver(){
	this.Status = models.IDLE
	this.Puid = ""
	this.P_x_scale = -1
	this.P_y_scale = -1
	this.D_x_scale = -1
	this.D_y_scale = -1

}


