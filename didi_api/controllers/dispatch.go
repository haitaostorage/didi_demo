package controllers

import (
	"log"
	"os"
	"strconv"
	"encoding/json"
	"sync"
	"didi_api/models"
	"github.com/gorilla/websocket"
	"github.com/astaxie/beego"
)

// Operations about order dispatch

type DispatchPoolController struct {
	beego.Controller
}

type Job struct{
	PayLoad interface{}
}



const	MAX_QUEUE int = 128

var (
	Logger = log.New(os.Stdout,"", log.LstdFlags)
	Dlock = sync.Mutex{}
	Olock = sync.Mutex{}
	Alock = sync.Mutex{}
	DriverList = make(map[string]*DriverElement,10)
	OrderList = make(map[string]*OrderElement,10)
	ArrangedDriver = make(map[string]*DriverElement,10)
	JobQueue = make(chan Job,MAX_QUEUE)
    )

func init(){
	scheduler := NewScheduler()
	scheduler.Run()
}


func (this *DispatchPoolController) JoinPool(){
	uid := this.GetString("uid")
	utype := this.GetString("type")
	x_scale := this.GetString("x_scale")
	y_scale := this.GetString("y_scale")
	d_x_scale := this.GetString("d_x_scale")
	d_y_scale := this.GetString("d_y_scale")
	if uid == "" || utype == ""{
		beego.Error(this.Ctx.ResponseWriter, "parameter error lack [uid or type]", 400)
		return
	}
	ws,err := websocket.Upgrade(this.Ctx.ResponseWriter,this.Ctx.Request,nil,1024,1024)
	if _, ok := err.(websocket.HandshakeError); ok{
		beego.Error(this.Ctx.ResponseWriter, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		beego.Error("Cannot setup WebSocket connection:", err)
		return
	}
	x,err1 := strconv.Atoi(x_scale)
	y,err2 := strconv.Atoi(y_scale)
	if err1 != nil || err2 != nil{
		ws.Close()
		beego.Error(this.Ctx.ResponseWriter, "invalid parameter[x_scale y_scale]", 400)
		return
	}
	if utype == "driver"{
		Logger.Println("driver  request")
		elem := &PoolElement{utype:"driver",uid:uid,op:"exist"}
		job := Job{PayLoad:elem}
		JobQueue<-job
		ret := <-DriverChan
		if ret{
			ws.Close()
			return
		}	
		//生成司机job结构,并放入待调度队列JobQueue
		driver := NewDriverElement(ws,uid,"",x,y,models.IDLE,-1,-1,-1,-1)	
		job = Job{PayLoad:driver}
		Logger.Println(driver)
		JobQueue <- job
	}else if utype == "passenger"{
		Logger.Println("passenger request")
		elem := &PoolElement{utype:"passenger",uid:uid,op:"exist"}
		job := Job{PayLoad:elem}
		JobQueue<-job
		ret := <-OrderChan
		if ret{
			ws.Close()
			return
		}
		var d_x int
		var d_y int
		if d_x_scale == "" || d_y_scale == ""{
			ws.Close()
			beego.Error(this.Ctx.ResponseWriter, "lock parameter[d_x_scale d_y_scale]", 400)
			return
		}		
		d_x,err1 = strconv.Atoi(d_x_scale)
		d_y,err2 = strconv.Atoi(d_y_scale)
		if err1 != nil || err2 != nil{
			ws.Close()
			beego.Error(this.Ctx.ResponseWriter, "invalid parameter[d_x_scale d_y_scale]", 400)
			return
		}
		 //向数据库中插入订单
		/*oid,err := AddOrderToDB(uid,x,y,d_x,d_y)   		
		if err != nil{
			ws.Close()
			http.Error(this.Ctx.ResponseWriter,err.Error(), 400)
			return
		}*/
		//生成订单job,放入JobQUEUE
		pass := NewOrderElement(ws,-1,uid,"",-1,-1,models.UNDISPATCH,x,y,d_x,d_y)
		job = Job{PayLoad:pass}
		JobQueue<-job
	}
}

//改进成为有调度协程统一进行池元素的添加和删除
func (this *DispatchPoolController) LeavePool(){
	uid := this.GetString("uid")
	utype := this.GetString("type")
	if utype == "driver"{
		elem,ok:= DriverList[uid]
		if ok{	
			//ResetDriver(elem)
			elem.Self_x_scale = -1
			elem.Self_y_scale = -1
			str ,_ := json.Marshal(elem)
			elem.Ws.WriteMessage(websocket.TextMessage, str)
			Leave(uid,utype)
		}//已接单的司机不许收车
	}

}

func Leave(uid string,utype string){
	if utype == "driver"{
		elem,ok := DriverList[uid]
		if ok{
			elem.Ws.Close()
			delete(DriverList,uid)
			return
		}
		elem,ok = ArrangedDriver[uid]
		if ok{
			elem.Ws.Close()
			delete(ArrangedDriver,uid)
			return
		}
	}else if utype == "passenger"{
		elem := OrderList[uid]
		if elem.Ws != nil{
			elem.Ws.Close()
		}
		delete(OrderList,uid)
	}else{
		return
	}
}


