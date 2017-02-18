package pool

//import	"github.com/donnie4w/go-logger/logger"

type PassengerState struct {
	uid    string
	x      int
	y      int
	status int
}

func NewPassengerState(uid string, x, y, status int) *PassengerState {
	return &PassengerState{uid, x, y, status}
}

func (this *PassengerState) Process() {
	this.UpdatePassengerState()
}

func (this *PassengerState) UpdatePassengerState() {
	Opool.Lock.Lock()
	v, ok := Opool.OrderList[this.uid]
	if ok {
		v.Self_x_scale = this.x
		v.Self_y_scale = this.y
		v.Status = this.status
		FlushOrderToCache(v.Puid,v)
	}
	Opool.Lock.Unlock()
}
