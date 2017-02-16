package main

import (
	"didi_api/models"
	"encoding/json"
	"github.com/donnie4w/go-logger/logger"
	"github.com/gorilla/websocket"
	"math/rand"
	"pool"
	"time"
)

func DriverPosition() {
	for {
		pool.Dpool.Lock.Lock()
		for _, v := range pool.Dpool.DriverList {
			x, y := RandomPosition(v.Self_x_scale, v.Self_y_scale)
			v.Self_x_scale = x
			v.Self_y_scale = y
			str, _ := json.Marshal(v)
			if v.Ws.WriteMessage(websocket.TextMessage, str) != nil {
				logger.Error("send driver positon infomation failure")
				pool.Leave(v.Uid, "driver")
			}
		}
		pool.Dpool.Lock.Unlock()
		time.Sleep(1e9)
	}
}

func ArrangeDriverPosition() {
	for {
		pool.Apool.Lock.Lock()
		for _, v := range pool.Apool.ArrangedList {
			pid := v.Puid
			x := v.Self_x_scale
			y := v.Self_y_scale
			p_x := v.P_x_scale
			p_y := v.P_y_scale
			d_x := v.D_x_scale
			d_y := v.D_y_scale
			var next_x, next_y int
			if (x != p_x || y != p_y) && v.Status == models.PREPARE { //接乘客过程
				next_x, next_y = NextPosition(x, y, p_x, p_y)
				v.Self_x_scale = next_x
				v.Self_y_scale = next_y
				state := pool.NewPassengerState(pid, next_x, next_y, models.DISPATCH)
				job := Job{PayLoad: state}
				JobQueue <- job
			} else if x == p_x && y == p_y && v.Status == models.PREPARE { //到达上车地点
				v.Status = models.READY
				if x != d_x || y != d_y {
					next_x, next_y = NextPosition(x, y, d_x, d_y)
					v.Self_x_scale = next_x
					v.Self_y_scale = next_y
					state := pool.NewPassengerState(pid, next_x, next_y, models.DISPATCH)
					job := Job{PayLoad: state}
					JobQueue <- job
				} else {
					state := pool.NewPassengerState(pid, next_x, next_y, models.COMPLETE)
					job := Job{PayLoad: state}
					JobQueue <- job
				}
			} else if v.Status == models.READY { //乘客上车，开始行程
				if x != d_x || y != d_y {
					next_x, next_y = NextPosition(x, y, d_x, d_y)
					v.Self_x_scale = next_x
					v.Self_y_scale = next_y
					state := pool.NewPassengerState(pid, next_x, next_y, models.DISPATCH)
					job := Job{PayLoad: state}
					JobQueue <- job
				} else { //到达目的地,将司机释放回调度池
					state := pool.NewPassengerState(pid, next_x, next_y, models.COMPLETE)
					job := Job{PayLoad: state}
					JobQueue <- job
					elem := pool.NewPoolElement("driver", "release", v.Uid)
					job = Job{PayLoad: elem}
					JobQueue <- job
					elem = pool.NewPoolElement("passenger", "release", pid)
					job = Job{PayLoad: elem}
					JobQueue <- job
				}

			}
			str, _ := json.Marshal(v)
			if v.Ws.WriteMessage(websocket.TextMessage, str) != nil {
				logger.Error("send driver positon infomation failure")
				pool.Leave(v.Uid, "driver")
			}
		}
		pool.Apool.Lock.Unlock()
		time.Sleep(1e9)
	}
}

func OrderPosition() {
	for {
		pool.Opool.Lock.Lock()
		for _, v := range pool.Opool.OrderList {
			str, _ := json.Marshal(v)
			if v.Ws.WriteMessage(websocket.TextMessage, str) != nil {
				logger.Error("send order infomation failure")
				pool.Leave(v.Puid, "passenger")
			}
		}
		pool.Opool.Lock.Unlock()
		time.Sleep(1e9)
	}
}

func RandomPosition(x, y int) (x_ret, y_ret int) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	num := r.Intn(4)
	if num == 0 {
		if x == 0 {
			x_ret = 1
			y_ret = y
		} else {
			x_ret = x - 1
			y_ret = y
		}
		return
	} else if num == 1 {
		if x == int(models.X_SCALE) {
			x_ret = int(models.X_SCALE) - 1
			y_ret = y
		} else {
			x_ret = x + 1
			y_ret = y
		}
		return
	} else if num == 2 {
		if y == 0 {
			x_ret = x
			y_ret = 1
		} else {
			x_ret = x
			y_ret = y - 1
		}
		return
	} else if num == 3 {
		if y == int(models.Y_SCALE) {
			x_ret = x
			y_ret = int(models.Y_SCALE) - 1
		} else {
			x_ret = x
			y_ret = y + 1
		}
		return
	}
	return
}

func NextPosition(x, y, d_x, d_y int) (rx, ry int) {
	if x < d_x {
		rx = x + 1
		ry = y
	} else if x > d_x {
		rx = x - 1
		ry = y
	} else if y < d_y {
		rx = x
		ry = y + 1
	} else if y > d_y {
		rx = x
		ry = y - 1
	} else {
		rx = x
		ry = y
	}
	return
}
