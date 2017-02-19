package pool

import (
	"didi_api/models"
	"encoding/json"
	"github.com/donnie4w/go-logger/logger"
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/websocket"
	"math"
	"strconv"
	"sync"
)

type OrderPool struct {
	Lock      sync.Mutex
	OrderList map[string]*OrderElement
}

type OrderElement struct {
	Ws           *websocket.Conn
	Uid          int64
	Puid         string
	Duid         string
	Self_x_scale int //司机位置
	Self_y_scale int //司机位置
	Status       int
	P_x_scale    int
	P_y_scale    int
	D_x_scale    int
	D_y_scale    int
}

var Opool *OrderPool

func NewOrderPool(size int) (err error){
	if RedisConn == nil {
		RedisConn, err = redis.Dial("tcp", "localhost:6379")
		if err != nil {
			logger.Error(err)
			return
		}
	}
	Opool = &OrderPool{sync.Mutex{}, make(map[string]*OrderElement, size)}
	reply, _ := redis.StringMap(RedisConn.Do("hgetall","order"))
	for k, v := range reply {
		var elem OrderElement
		json.Unmarshal([]byte(v),&elem)
		elem.Ws = nil
		Opool.OrderList[k] = &elem
	}
	return nil
}

func NewOrder(pid, x, y, d_x, d_y int) models.Orders {
	o := models.Orders{
		Passenger_id:  pid,
		Start_x_scale: x,
		Start_y_scale: y,
		End_x_scale:   d_x,
		End_y_scale:   d_y,
		Status:        models.UNDISPATCH}
	return o
}

func AddOrderToDB(uid string, x, y, d_x, d_y int) (oid int64, err error) {
	pid, err1 := strconv.Atoi(uid)
	if err1 != nil {
		err = err1
		return
	}
	porder := NewOrder(pid, x, y, d_x, d_y)
	oid, err = models.AddOrder(porder) //生成数据库订单
	return
}

func NewOrderElement(ws *websocket.Conn, oid int64, uid, id string, x, y, status, p_x, p_y, d_x, d_y int) *OrderElement {
	oi := OrderElement{
		Ws:           ws,
		Uid:          oid,
		Puid:         uid,
		Duid:         id,
		Self_x_scale: x,
		Self_y_scale: y,
		Status:       status,
		P_x_scale:    p_x,
		P_y_scale:    p_y,
		D_x_scale:    d_x,
		D_y_scale:    d_y}
	return &oi
}

func (this *OrderElement) Process() {
	this.JoinOrderPool()
	this.ArrangeDriver()
}

func (this *OrderElement) JoinOrderPool() {
	oid, err := AddOrderToDB(this.Puid, this.P_x_scale, this.P_y_scale, this.D_x_scale, this.D_y_scale)
	logger.Info("insert an order into database")
	if err != nil {
		logger.Error("insert an order into database error", err)
		this.Ws.Close()
		return
	}
	this.Uid = oid
	logger.Debug("an order join in pool")
	Opool.AddOrder(this)
}

func (this *OrderPool) UpdateOrderDriverInfo(e *OrderElement, uid string) {
	this.Lock.Lock()
	e.Duid = uid
	e.Status = models.DISPATCH
	this.Lock.Unlock()
	FlushOrderToCache(e)
}

func (this *OrderPool) AddOrder(e *OrderElement) {
	this.Lock.Lock()
	this.OrderList[e.Puid] = e
	this.Lock.Unlock()
	FlushOrderToCache(e)

}

func (this *OrderPool) DelOrder(e *OrderElement) *OrderElement {
	this.Lock.Lock()
	elem := this.OrderList[e.Puid]
	delete(this.OrderList, e.Puid)
	this.Lock.Unlock()
	DelOrderFromCache(e.Puid)
	return elem
}

func (this *OrderPool) DelOrderById(uid string) *OrderElement {
	this.Lock.Lock()
	elem := this.OrderList[uid]
	delete(this.OrderList, uid)
	this.Lock.Unlock()
	DelOrderFromCache(uid)
	return elem
}

func (this *OrderElement) ResetOrder() {
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

func (this *OrderElement) ArrangeDriver() {
	var near float64 = models.X_SCALE + models.Y_SCALE + 1
	var dselect *DriverElement = nil
	for _, v := range Dpool.DriverList {
		x := v.Self_x_scale
		y := v.Self_y_scale
		distance := math.Abs(float64(x + y - this.P_x_scale - this.P_y_scale))
		if distance < near {
			near = distance
			dselect = v
		}
	}
	if dselect == nil {
		logger.Warn("not find an suitable driver for the order")
		return
	}
	Dpool.UpdateDriverOrderInfo(dselect, this)     //更新司机的信息
	Dpool.DelDriver(dselect)
	Opool.UpdateOrderDriverInfo(this, dselect.Uid) //更新订单信息
	Apool.AddArrangedDriver(dselect)               //添加到已分配司机池
	logger.Info("arrange an suitbale driver for the order")
}

func FlushOrderToCache(elem *OrderElement) {
	str, err := json.Marshal(elem)
	if err != nil {
		logger.Error("FlushOrderToCache failure", err)
		return
	}
	_, err = RedisConn.Do("hset", "order", elem.Puid, str)
	if err != nil {
		logger.Error("FlushOrderToCache failure", err)
	}
	return

}

func DelOrderFromCache(key string) {
	_, err := RedisConn.Do("hdel", "order", key)
	if err != nil {
		logger.Error("DelOrderFromCache failure", err)
	}
}
