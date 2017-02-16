package controllers

import (
	"time"
	"math/rand"
	"encoding/json"
	"didi_api/models"
	"github.com/gorilla/websocket"
)


func init(){
	go DriverPosition()   //负责空闲司机位置协程
	go ArrangeDriverPosition()  //负责已接单司机位置
	go OrderPosition()    //负责乘客位置
}


func DriverPosition(){
	for{
		Dlock.Lock()
		for _,v := range DriverList{
			x,y := RandomPosition(v.Self_x_scale,v.Self_y_scale)
			v.Self_x_scale = x
			v.Self_y_scale = y
			str ,_ := json.Marshal(v)
			if v.Ws.WriteMessage(websocket.TextMessage, str) != nil{
				Leave(v.Uid,"driver")
			}
		}
		Dlock.Unlock()
		time.Sleep(1e9)
	}
}

func ArrangeDriverPosition(){
	for{
		Alock.Lock()
		for _,v := range ArrangedDriver{
			pid := v.Puid
			x := v.Self_x_scale
			y := v.Self_y_scale
			p_x := v.P_x_scale	
			p_y := v.P_y_scale
			d_x := v.D_x_scale 
			d_y := v.D_y_scale
			var next_x,next_y int
			if (x != p_x || y != p_y) && v.Status == models.PREPARE{//接乘客过程
				next_x,next_y = NextPosition(x,y,p_x,p_y)
				v.Self_x_scale = next_x
				v.Self_y_scale = next_y
				state := &PassengerState{pid,next_x,next_y,models.DISPATCH}
				job := Job{PayLoad:state}
				JobQueue<-job
			}else if x == p_x && y == p_y && v.Status == models.PREPARE{//到达上车地点
				v.Status = models.READY
				if x != d_x || y != d_y{
					next_x,next_y = NextPosition(x,y,d_x,d_y)
					v.Self_x_scale = next_x
					v.Self_y_scale = next_y
					state := &PassengerState{pid,next_x,next_y,models.DISPATCH}
					job := Job{PayLoad:state}
					JobQueue<-job
				}else{
					state := &PassengerState{pid,next_x,next_y,models.COMPLETE}
					job := Job{PayLoad:state}
					JobQueue<-job
				}
			}else if v.Status == models.READY{   //乘客上车，开始行程
				if x != d_x || y != d_y{
					next_x,next_y = NextPosition(x,y,d_x,d_y)
					v.Self_x_scale = next_x
					v.Self_y_scale = next_y
					state := &PassengerState{pid,next_x,next_y,models.DISPATCH}
					job := Job{PayLoad:state}
					JobQueue<-job
				}else{                          //到达目的地,将司机释放回调度池
					state := &PassengerState{pid,next_x,next_y,models.COMPLETE}
					job := Job{PayLoad:state}
					JobQueue<-job
					elem := &PoolElement{utype:"driver",uid:v.Uid,op:"release"}
					job = Job{PayLoad:elem}
					JobQueue<-job
					elem = &PoolElement{utype:"passenger",uid:pid,op:"release"}
					job = Job{PayLoad:elem}
					JobQueue<-job
				}

			} 
			str,_ := json.Marshal(v)
			if v.Ws.WriteMessage(websocket.TextMessage, str) != nil{
				Leave(v.Uid,"driver")
			}
		}
		Alock.Unlock()
		time.Sleep(1e9)
	}
}

func OrderPosition(){
	for{
		Olock.Lock()
		for _,v := range OrderList{
			str,_:= json.Marshal(v)
			if v.Ws.WriteMessage(websocket.TextMessage, str) != nil{
				Leave(v.Puid,"passenger")
			}
		}
		Olock.Unlock()
		time.Sleep(1e9)
	}
}

func RandomPosition(x,y int)(x_ret,y_ret int){
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	num := r.Intn(4)
	if num == 0{
		if x==0{
			x_ret = 1
			y_ret = y
		}else{
			x_ret = x - 1
			y_ret = y
		}
		return
	}else if num == 1{
		if x == int(models.X_SCALE){
			x_ret = int(models.X_SCALE) - 1
			y_ret = y
		}else{
			x_ret = x + 1
			y_ret = y
		}
		return
	}else if num == 2{
		if y==0{
			x_ret = x
			y_ret = 1
		}else{
			x_ret = x 
			y_ret = y - 1
		}
		return
	}else if num == 3 {
		if y == int(models.Y_SCALE){
			x_ret = x
			y_ret = int(models.Y_SCALE) - 1
		}else{
			x_ret = x
			y_ret = y + 1
		}
		return
	}
	return
}

func NextPosition(x,y,d_x,d_y int)(rx,ry int){
	if x < d_x{
		rx = x + 1
		ry = y
	}else if x > d_x{
		rx = x - 1
		ry = y
	}else if y < d_y{
		rx = x
		ry = y + 1
	}else if y > d_y{
		rx = x
		ry = y - 1
	}else{
		rx = x
		ry = y
	}
	return
}


