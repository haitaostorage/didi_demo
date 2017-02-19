package pool

import(
	"didi_api/models"
	"encoding/json"
	"github.com/donnie4w/go-logger/logger"
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/websocket"
	"sync"
)

type DriverPool struct {
	Lock	   sync.Mutex
	DriverList map[string]*DriverElement
}

type ArrangedDriverPool struct {
	Lock	     sync.Mutex
	ArrangedList map[string]*DriverElement
}

var Dpool *DriverPool
var Apool *ArrangedDriverPool
var RedisConn redis.Conn = nil

type DriverElement struct {
	Ws	     *websocket.Conn
	Uid	     string
	Puid	     string    //该司机接送的乘客id
	Self_x_scale int
	Self_y_scale int
	Status	     int
	P_x_scale    int
	P_y_scale    int
	D_x_scale    int
	D_y_scale    int
}

func NewDriverPool(size int) (err error) {
	if RedisConn == nil {
		RedisConn, err = redis.Dial("tcp", "localhost:6379")
		if err != nil {
			logger.Error(err)
			return
		}
	}
	Dpool = &DriverPool{sync.Mutex{}, make(map[string]*DriverElement, size)}
	reply, _ := redis.StringMap(RedisConn.Do("hgetall","driver"))
	for k, v := range reply {
		var elem DriverElement
		json.Unmarshal([]byte(v),&elem)
		elem.Ws = nil
		Dpool.DriverList[k] = &elem
	}
	return nil
}

func NewArrangedDriverPool(size int) (err error) {
	if RedisConn == nil {
		RedisConn, err = redis.Dial("tcp", "localhost:6379")
		if err != nil {
			logger.Error(err)
			return
		}
	}
	Apool =  &ArrangedDriverPool{sync.Mutex{}, make(map[string]*DriverElement, size)}
	reply, _ := redis.StringMap(RedisConn.Do("hgetall","adriver"))
	for k, v := range reply {
		var elem DriverElement
		json.Unmarshal([]byte(v),&elem)
		elem.Ws = nil
		Apool.ArrangedList[k] = &elem
	}
	return nil
}

func NewDriverElement(ws *websocket.Conn, uid, pid string, x, y, status, p_x, p_y, d_x, d_y int) *DriverElement {
	e := DriverElement{ws, uid, pid, x, y, status, p_x, p_y, d_x, d_y}
	return &e
}

func (this *DriverElement) Process() {
	this.JoinDriverPool()
	Sched()
}

func (this *DriverElement) JoinDriverPool() {
	logger.Debug("an driver join in the pool")
	Dpool.AddDriver(this)
}

func (this *DriverPool) AddDriver(e *DriverElement) {
	this.Lock.Lock()
	this.DriverList[e.Uid] = e
	this.Lock.Unlock()
	FlushDriverToCache(e)

}

func (this *DriverPool) DelDriver(e *DriverElement) *DriverElement {
	this.Lock.Lock()
	elem := this.DriverList[e.Uid]
	delete(this.DriverList, e.Uid)
	this.Lock.Unlock()
	DelDriverFromCache(e.Uid)
	return elem
}

func (this *DriverPool) DelDriverById(uid string) *DriverElement {
	this.Lock.Lock()
	elem := this.DriverList[uid]
	delete(this.DriverList, uid)
	this.Lock.Unlock()
	DelDriverFromCache(uid)
	return elem
}


func (this *DriverPool) UpdateDriverOrderInfo(e *DriverElement, orderinfo *OrderElement) {
	this.Lock.Lock()
	e.Puid = orderinfo.Puid
	e.Status = models.PREPARE
	e.P_x_scale = orderinfo.P_x_scale
	e.P_y_scale = orderinfo.P_y_scale
	e.D_x_scale = orderinfo.D_x_scale
	e.D_y_scale = orderinfo.D_y_scale
	this.Lock.Unlock()
	FlushDriverToCache(e)
}

func (this *ArrangedDriverPool) AddArrangedDriver(e *DriverElement) {
	this.Lock.Lock()
	this.ArrangedList[e.Uid] = e
	this.Lock.Unlock()
	FlushAdriverToCache(e)
}

func (this *ArrangedDriverPool) DelArrangedDriver(e *DriverElement) *DriverElement {
	elem := this.ArrangedList[e.Uid]
	this.Lock.Lock()
	delete(this.ArrangedList, e.Uid)
	this.Lock.Unlock()
	DelAdriverFromCache(e.Uid)
	return elem
}

func (this *ArrangedDriverPool) DelArrangedDriverById(uid string) *DriverElement {
	this.Lock.Lock()
	elem := this.ArrangedList[uid]
	delete(this.ArrangedList, uid)
	this.Lock.Unlock()
	DelAdriverFromCache(uid)
	return elem
}


func (this *DriverElement) Reset() {
	this.Status = models.IDLE
	this.Puid = ""
	this.P_x_scale = -1
	this.P_y_scale = -1
	this.D_x_scale = -1
	this.D_y_scale = -1
}

func Sched() {
	for _, v := range Opool.OrderList {
		if v.Status == models.UNDISPATCH {
			//Logger.Println("arrange driver for",v.Puid)
			v.ArrangeDriver()
		}
	}
}

func FlushDriverToCache(elem *DriverElement) {
	str, err := json.Marshal(elem)
	if err != nil {
		logger.Error("FlushDriverToCache error:", err)
		return
	}
	_, err = RedisConn.Do("hset", "driver", elem.Uid, str)
	if err != nil {
		logger.Error("FlushDriverToCache failure", err)
	}
	return
}

func FlushAdriverToCache(elem *DriverElement) {
	str, err := json.Marshal(elem)
	if err != nil {
		logger.Error("FlushAdriverToCache error:", err)
		return
	}
	_, err = RedisConn.Do("hset", "adriver", elem.Uid, str)
	if err != nil {
		logger.Error("FlushAdriverToCache failure", err)
	}
	return
}

func DelDriverFromCache(key string) {
	_, err := RedisConn.Do("hdel", "driver", key)
	if err != nil {
		logger.Error("DelDriverFromCache failure", err)
	}
}

func DelAdriverFromCache(key string) {
	_, err := RedisConn.Do("hdel", "adriver", key)
	if err != nil {
		logger.Error("DelAdriverFromCache failure", err)
	}
}
