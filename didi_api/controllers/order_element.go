package controllers
import(
	"github.com/gorilla/websocket"
	"didi_api/models"
	"strconv"
	"math"
)

type OrderElement struct{
	Ws *websocket.Conn
	Uid int64
	Puid string
	Duid string
	Self_x_scale int   //司机位置
	Self_y_scale int   //司机位置
	Status int
	P_x_scale int
	P_y_scale int
	D_x_scale int
	D_y_scale int
}

func NewOrder(pid,x,y,d_x,d_y int)models.Orders{
	o :=  models.Orders{
		Passenger_id:pid,
		Start_x_scale:x,
		Start_y_scale:y,
		End_x_scale:d_x,
		End_y_scale:d_y,
		Status:models.UNDISPATCH}
	return o
}

func AddOrderToDB(uid string,x,y,d_x,d_y int)(oid int64,err error){
	pid,err1 := strconv.Atoi(uid)
	if err1 != nil{
		err = err1	
		return
	}
	porder := NewOrder(pid,x,y,d_x,d_y)
	oid,err = models.AddOrder(porder)    //生成数据库订单
	return
}


func NewOrderElement(ws *websocket.Conn,oid int64,uid,id string,x,y,status,p_x,p_y,d_x,d_y int)*OrderElement{
	oi := OrderElement{
		Ws:ws,
		Uid:oid,
		Puid:uid,
		Duid:id,
		Self_x_scale:x,
		Self_y_scale:y,
		Status:status,
		P_x_scale:p_x,
		P_y_scale:p_y,
		D_x_scale:d_x,
		D_y_scale:d_y}
	return &oi
}

func (this *OrderElement)JoinOrderPool(){
	oid,err := AddOrderToDB(this.Puid,this.P_x_scale,this.P_y_scale,this.D_x_scale,this.D_y_scale)
	if err != nil{
		this.Ws.Close()
		return
	}
	this.Uid = oid
	Olock.Lock()
	OrderList[this.Puid] = this
	Olock.Unlock()
}

func (this *OrderElement)UpdateOrderDriverInfo(uid string){
	this.Duid = uid
	this.Status = models.DISPATCH
}

func (this *OrderElement)ResetOrder(){
	this.Status = models.UNDISPATCH
	this.Puid = ""
	this.Duid = ""
	this.Self_x_scale = -1
	this.Self_y_scale = -1
	this.P_x_scale = -1
	this.P_y_scale = -1
	this.D_x_scale = -1
	this.D_y_scale = -1

}

func (this *OrderElement) ArrangeDriver(){
	Dlock.Lock()
	var near float64 = models.X_SCALE + models.Y_SCALE + 1
	var dselect *DriverElement = nil
	var dselect_k string
	for k,v := range DriverList{
		x := v.Self_x_scale
		y := v.Self_y_scale
		distance := math.Abs(float64(x + y - this.P_x_scale - this.P_y_scale))
		if distance < near{
			near = distance
			dselect = v
			dselect_k = k
		}
	}
	if dselect == nil{
		Dlock.Unlock()
		return
	}
	delete(DriverList,dselect_k)
	Dlock.Unlock()
	Olock.Lock()
	this.UpdateOrderDriverInfo(dselect.Uid)  //更新订单信息
	dselect.UpdateDriverOrderInfo(this)     //更新司机的信息
	Olock.Unlock()
	Alock.Lock()
	ArrangedDriver[dselect_k] = dselect
	Alock.Unlock()
}

