package models

import (
	"math"
	"errors"
	"strconv"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"

)

const X_SCALE float64 = 200		//坐标系X轴范围
const Y_SCALE float64 = 200	 	//坐标系Y轴范围

const(
	OFFLINE = iota
	ONLINE
     )

const(
	IDLE = iota
	PREPARE
	READY
     )

type Driver struct {
	Id       int `orm:"auto"`
	Name string `orm:"size(100)"`
	X_scale	int
	Y_scale int
	Balance	int	//司机钱包余额
	Online	int	//0:收车 1:出车
	Status	int	//online=1有效 0:空闲，可以接单 1:接单，还未接客 2：行程中

}

var ConsoleLogs *logs.BeeLogger
func init() {
	orm.RegisterModel(new(Driver))
	 ConsoleLogs = logs.NewLogger(1000)
	 ConsoleLogs.SetLogger("console")
}


func AddDriver(u Driver)(id int64,err error) {
	o := orm.NewOrm()
	id,err1 := o.Insert(&u)
	if err1 != nil{
		err = err1
		return
	}
	return id,nil
}

func GetAllDrivers()(m []orm.Params,err error){
	var maps []orm.Params
	o := orm.NewOrm()
	_,err1 := o.QueryTable("driver").Values(&maps)
	if err1 != nil{
		err = err1
		return
	}else{
		return maps,nil
	}
}

func GetDriver(uid string) (driver *Driver,err error){
	o := orm.NewOrm()
	id,err1 := strconv.Atoi(uid)
	if err1!=nil{
		err = err1
		return
	}
	d := Driver{Id:id}
	err1 = o.Read(&d)
	if err1 != nil{
		err = err1
		return
	}else{
		return &d,nil
	}
}

func UpdateDriver(uid string,dd *Driver) (driver *Driver,err error){
	o := orm.NewOrm()
	id,err1 := strconv.Atoi(uid)
	if err1 != nil{
		err = err1
		return
	}
	d := Driver{Id:id}
	err1 = o.Read(&d)
	if err1 != nil{
		err = err1
		return
	}else{
		d.Online = dd.Online
		d.Status = dd.Status
		d.Balance = dd.Balance
		d.X_scale = dd.X_scale
		d.Y_scale = dd.Y_scale
		_,err2 := o.Update(&d)
		if err2 != nil{
			err = err2
			return
		}
		return &d,nil
	}
}

func DeleteDriver(uid string) (err error){
	o := orm.NewOrm()
	id,err1 := strconv.Atoi(uid)
	if err1 != nil{
		err = err1
		return
	}
	d := Driver{Id:id}
	_,err2 := o.Delete(&d)
	if err2 != nil{
		err = err2
		return
	}
	return nil
}

func Charge(uid string,dd *Driver)(driver *Driver,err error){
	o := orm.NewOrm()
	id,err1 := strconv.Atoi(uid)
	if err1 != nil{
		err = err1
		return
	}
	d := Driver{Id:id}
	err1 = o.Read(&d)
	if err1 != nil{
		err = err1
		return
	}else{
		d.Balance = d.Balance + dd.Balance
		_,err2 := o.Update(&d)
		if err2 != nil{
			err = err2
			return
		}
		return &d,nil
	}
	return &d,nil
}

func Dispatch(uid string)(d *Driver,err error){
	o := orm.NewOrm()
	id,err1 := strconv.Atoi(uid)
	if err1 != nil{
		err = err1
		return
	}
	order := Orders{Id:id}
	err1 = o.Read(&order)
	if err1 != nil{
		err = err1
		return
	}
	var maps []orm.Params
	_,err2 := o.QueryTable("driver").Filter("Online",ONLINE).Filter("Status",IDLE).Values(&maps)
	//qs.Filter("Online",1).Filter("Status",0)
	//_,err2 := qs.Values(&maps,"id","name","x_scale","y_scale","balance","online")
	//fmt.Println(maps)
	if err2 != nil{
		err = err2
		return
	}
	var near float64 = X_SCALE * Y_SCALE
	var dselect orm.Params
	for _,m := range maps{
		distance := math.Abs(float64(m["X_scale"].(int64) + m["Y_scale"].(int64) -int64(order.Start_x_scale) - int64(order.Start_y_scale)))
		if distance < near{
			near = distance
			dselect = m
		}
	}
	if near == X_SCALE * Y_SCALE{
		err = errors.New("not find suitable driver")
		return
	}
	driver := Driver{
			Id : int(dselect["Id"].(int64)),
			Name : dselect["Name"].(string),
			X_scale : int(dselect["X_scale"].(int64)),
			Y_scale : int(dselect["Y_scale"].(int64)),
			Online  : int(dselect["Online"].(int64)),
			Balance : int(dselect["Balance"].(int64)),
			Status  : PREPARE,
		  }
	_,err3 := o.Update(&driver)        //找到合适的司机接单后更新司机状态
	if err3 != nil{
		err = err3
		return
	}
	order.Driver_id = driver.Id
	order.Status = DISPATCH
	_,err4 := o.Update(&order)	//找到合适的司机接单后更新订单状态
	if err4 != nil{
		err = err4
		return
	}
	return &driver,nil
}
