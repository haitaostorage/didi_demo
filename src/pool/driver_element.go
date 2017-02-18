package pool

import(
	"didi_api/models"
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

func NewDriverPool(size int) err Error {
	if RedisConn == nil {
		RedisConn, err1 := redis.Dail("tcp", "localhost:6379")
		if err1 != nil {
			err = err1
			logger.Error(err)
			return
		}
	}
	Dpool = &DriverPool{sync.Mutex{}, make(map[string]*DriverElement, size)}
	reply, err1 := redis.StringMap(RedisConn.Do("hgetall","driver"))
	for k, v := range reply {
		var elem Driver
		json.Unmarshal([]byte(v),&elem)
		elem.Ws = nil
		Dpool.DriverList[k] = elem
	}
}

func NewArrangedDriverPool(size int) {
	if RedisConn == nil {
		RedisConn, err1 := redis.Dail("tcp", "localhost:6379")
		if err1 != nil {
			err = err1
			logger.Error(err)
			return
		}
	}
	Apool =  &ArrangedDriverPool{sync.Mutex{}, make(map[string]*DriverElement, size)}
	reply, err1 := redis.StringMap(RedisConn.Do("hgetall","adriver"))
	for k, v := range reply {
		var elem ArrangedDriver
		json.Unmarshal([]byte(v),&elem)
		elem.Ws = nil
		Apool.ArrangedList[k] = elem
	}
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

}

func (this *DriverPool) DelDriver(e *DriverElement) *DriverElement {
	this.Lock.Lock()
	elem := this.DriverList[e.Uid]
	delete(this.DriverList, e.Uid)
	this.Lock.Unlock()
	return elem
}

func (this *DriverPool) DelDriverById(uid string) *DriverElement {
	this.Lock.Lock()
	elem := this.DriverList[uid]
	delete(this.DriverList, uid)
	this.Lock.Unlock()
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
}

func (this *ArrangedDriverPool) AddArrangedDriver(e *DriverElement) {
	this.Lock.Lock()
	this.ArrangedList[e.Uid] = e
	this.Lock.Unlock()
}

func (this *ArrangedDriverPool) DelArrangedDriver(e *DriverElement) *DriverElement {
	elem := this.ArrangedList[e.Uid]
	this.Lock.Lock()
	delete(this.ArrangedList, e.Uid)
	this.Lock.Unlock()
	return elem
}

func (this *ArrangedDriverPool) DelArrangedDriverById(uid string) *DriverElement {
	this.Lock.Lock()
	elem := this.ArrangedList[uid]
	delete(this.ArrangedList, uid)
	this.Lock.Unlock()
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

func FlushDriverToCache(key string, elem *DriverElement) {
	str, err := json.Marshal(elem)
	if err != nil {
		logger.Error("FlushDriverToCache error:", err)
		return
	}
	_, err = RedisConn.Do("hset", "driver", key, str)
	if err != nil {
		logger.Error("FlushDriverToCache error", err)
	}
	return
}

func FlushAdriverToCache(key string, elem *DriverElement) {
	str, err := json.Marshal(elem)
	if err != nil {
		logger.Error("FlushAdriverToCache error:", err)
		return
	}
	_, err = RedisConn.Do("hset", "adriver", key, str)
	if err != nil {
		logger.Error("FlushAdriverToCache error", err)
	}
	return
}
