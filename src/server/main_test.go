package main

import (
	"github.com/gorilla/websocket"
	"strconv"
	"testing"
)

var conn []*websocket.Conn = make([]*websocket.Conn, 5000)
var dconn []*websocket.Conn = make([]*websocket.Conn, 5000)

func Benchmark_OrderJoin(b *testing.B) {
	for i := 0; i < b.N; i++ {
		path := "ws://127.0.0.1:80/v1/ds?uid=" + strconv.Itoa(i) + "&type=passenger&x_scale=0&y_scale=0&d_x_scale=0&d_y_scale=0"
		var err error
		conn[i], _, err = websocket.DefaultDialer.Dial(path, nil)
		if err != nil {
			b.Errorf("websocket request failure:", err.Error())
		}
	}
}

func Benchmark_DriverJoin(b *testing.B) {
	for i := 0; i < b.N; i++ {
		path := "ws://127.0.0.1:80/v1/ds?uid=" + strconv.Itoa(i) + "&type=driver&x_scale=0&y_scale=0&d_x_scale=0&d_y_scale=0"
		var err error
		dconn[i], _, err = websocket.DefaultDialer.Dial(path, nil)
		if err != nil {
			b.Errorf("websocket request failure:", err.Error())
		}
	}

}
