package models

import (
	"strconv"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"

)

const(
	UNDISPATCH = iota
	DISPATCH
	COMPLETE
	CHARGED
     )

type Orders struct {
	Id       int `orm:"auto"`
	Passenger_id int
	Driver_id    int
	Start_x_scale	int
	Start_y_scale 	int
	End_x_scale	int
	End_y_scale	int
	Status		int//0: 未被分配  1：已分配，进行中 2：行程结束，未支付 3：行程结束，已支付

}


func init() {
	orm.RegisterModel(new(Orders))

}


func AddOrder(u Orders)(id int64,err error) {
	o := orm.NewOrm()
	id,err1 := o.Insert(&u)
	if err1 != nil{
		err = err1
		return
	}
	return id,nil
}
func GetAllOrders()(m []orm.Params,err error){
	var maps []orm.Params
	o := orm.NewOrm()
	_,err1 := o.QueryTable("orders").Values(&maps)
	if err1 != nil{
		err = err1
		return
	}else{
		return maps,nil
	}
}

func GetOrder(uid string) (order *Orders,err error){
	o := orm.NewOrm()
	id,err1 := strconv.Atoi(uid)
	if err1!=nil{
		err = err1
		return
	}
	d := Orders{Id:id}
	err1 = o.Read(&d)
	if err1 != nil{
		err = err1
		return
	}else{
		return &d,nil
	}
}

func UpdateOrder(uid string,dd *Orders) (order *Orders,err error){
	o := orm.NewOrm()
	id,err1 := strconv.Atoi(uid)
	if err1 != nil{
		err = err1
		return
	}
	d := Orders{Id:id}
	err1 = o.Read(&d)
	if err1 != nil{
		err = err1
		return
	}else{
		d.Driver_id = dd.Driver_id
		d.Status = dd.Status
		d.Start_x_scale = dd.Start_x_scale
		d.Start_y_scale = dd.Start_y_scale
		d.End_x_scale = dd.End_x_scale
		d.End_y_scale = dd.End_y_scale
		_,err2 := o.Update(&d)
		if err2 != nil{
			err = err2
			return
		}
		return &d,nil
	}
}

func DeleteOrder(uid string) (err error){
	o := orm.NewOrm()
	id,err1 := strconv.Atoi(uid)
	if err1 != nil{
		err = err1
		return
	}
	d := Orders{Id:id}
	_,err2 := o.Delete(&d)
	if err2 != nil{
		err = err2
		return
	}
	return nil
}
