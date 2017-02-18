package main

import (
	"pool"
	//"os"
	"didi_api/models"
	"encoding/json"
	"github.com/astaxie/beego/orm"
	"github.com/donnie4w/go-logger/logger"
	"github.com/garyburd/redigo/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
	"net/http"
	"strconv"
)

type Job struct {
	PayLoad interface{}
}

const MAX_QUEUE int = 1
const DRIVER_POOL_SIZE int = 1024
const ORDER_POOL_SIZE int = 1024

var (
	JobQueue = make(chan Job, MAX_QUEUE)
)

func init_log() {
	logger.SetConsole(false)
	logger.SetRollingDaily("/root/didi_demo/log", "server.log")
	logger.SetLevel(logger.ERROR)
}

func position_routine() {
	go DriverPosition()        //负责空闲司机位置协程
	go ArrangeDriverPosition() //负责已接单司机位置
	go OrderPosition()         //负责乘客位置
}

func main() {
	//初始化日志
	init_log()
	logger.Info("initialize logger")

	//初始化订单和司机池
	pool.NewDriverPool(DRIVER_POOL_SIZE)
	pool.NewArrangedDriverPool(DRIVER_POOL_SIZE)
	pool.NewOrderPool(ORDER_POOL_SIZE)
	logger.Info("initialize driver and order pool")

	//启动后台调度器
	scheduler := NewScheduler()
	scheduler.Run()
	logger.Info("initialize scheduler and worker")

	//启动位置协程
	position_routine()
	logger.Info("initialize position routine")

	//初始化数据库连接
	orm.RegisterDriver("mysql", orm.DRMySQL)
	orm.RegisterDataBase("default", "mysql", "root:123456@tcp(127.0.0.1:3306)/didi_test?charset=utf8", 30)
	orm.RunSyncdb("default", false, true)
	logger.Info("initialize database connection")

	//启动http服务器
	http.HandleFunc("/v1/ds", JoinPool)
	http.HandleFunc("/v1/leave", LeavePool)
	logger.Info("start http server and waiting for connection......")
	err := http.ListenAndServe("0.0.0.0:80", nil)
	if err != nil {
		logger.Fatal(err)
	}
}

func JoinPool(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if len(r.Form) < 4 {
		logger.Error("lack parameters")
		return
	}
	uid := r.Form.Get("uid")
	utype := r.Form.Get("type")
	x_scale := r.Form.Get("x_scale")
	y_scale := r.Form.Get("y_scale")
	d_x_scale := r.Form.Get("d_x_scale")
	d_y_scale := r.Form.Get("d_y_scale")
	if uid == "" || utype == "" {
		logger.Error("parameter error lack uid or type")
		return
	}
	ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		logger.Error("Not a websocket handshake")
		return
	} else if err != nil {
		logger.Error("Cannot setup WebSocket connection:", err)
		return
	}
	x, err1 := strconv.Atoi(x_scale)
	y, err2 := strconv.Atoi(y_scale)
	if err1 != nil || err2 != nil {
		ws.Close()
		logger.Error("invalid parameter x_scale y_scale")
		return
	}
	if utype == "driver" {
		logger.Info("driver online request")
		elem := pool.NewPoolElement("driver", "exist", uid)
		job := Job{PayLoad: elem}
		JobQueue <- job
		ret := <-pool.DriverChan
		if ret {
			logger.Warn("driver already exist in pool")
			ws.Close()
			return
		}
		//生成司机job结构,并放入待调度队列JobQueue
		driver := pool.NewDriverElement(ws, uid, "", x, y, models.IDLE, -1, -1, -1, -1)
		job = Job{PayLoad: driver}
		JobQueue <- job
	} else if utype == "passenger" {
		logger.Info("passenger create order request")
		elem := pool.NewPoolElement("passenger", "exist", uid)
		job := Job{PayLoad: elem}
		JobQueue <- job
		ret := <-pool.OrderChan
		if ret {
			logger.Warn("passenger already create an order")
			ws.Close()
			return
		}
		var d_x int
		var d_y int
		if d_x_scale == "" || d_y_scale == "" {
			ws.Close()
			logger.Error("lack parameter d_x_scale or d_y_scale")
			return
		}
		d_x, err1 = strconv.Atoi(d_x_scale)
		d_y, err2 = strconv.Atoi(d_y_scale)
		if err1 != nil || err2 != nil {
			ws.Close()
			logger.Error("invalid parameter d_x_scale d_y_scale")
			return
		}
		//生成订单job,放入JobQUEUE
		pass := pool.NewOrderElement(ws, -1, uid, "", -1, -1, models.UNDISPATCH, x, y, d_x, d_y)
		job = Job{PayLoad: pass}
		JobQueue <- job
	}
}

//改进成为有调度协程统一进行池元素的添加和删除
func LeavePool(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	uid := r.Form.Get("uid")
	utype := r.Form.Get("type")
	if utype == "driver" {
		elem, ok := pool.Dpool.DriverList[uid]
		if ok {
			//ResetDriver(elem)
			elem.Self_x_scale = -1
			elem.Self_y_scale = -1
			str, _ := json.Marshal(elem)
			elem.Ws.WriteMessage(websocket.TextMessage, str)
			pool.Leave(uid, utype)
		} //已接单的司机不许收车
	}

}
