package models

import (
	"strconv"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

const (
	P_IDLE = iota
	P_PROCESSING
	P_COMPLETE
      )

type Passenger struct {
	Id       int	`orm:"auto"`
	Name string `orm:"size(100)"`
	Status 	 int   //0 :空闲  1: 行程中 2: 行程结束未支付
}


func init() {
	orm.RegisterModel(new(Passenger))
}

func AddPassenger(u Passenger) (id int64,err error){
	o := orm.NewOrm()
	id,err1 := o.Insert(&u)
	if err1!=nil{
		err = err1
		return
	}
	return id,nil
}

func GetPassenger(id string) (u *Passenger, err error) {
	o := orm.NewOrm()
	uid,err1 := strconv.Atoi(id)
	if err1!=nil{
		err = err1
		return
	}
	p := Passenger{Id:uid}
	err = o.Read(&p)
	return &p,err
}

func GetAllPassengers() (list []orm.Params,err error) {
	var maps []orm.Params
	o := orm.NewOrm()
	_,err1 := o.QueryTable("passenger").Values(&maps)
	if err1 !=nil{
		err = err1
		return
	}else{
		return maps,nil
	}
}

func UpdatePassenger(id string,passenger *Passenger) (u *Passenger,err error){
	o := orm.NewOrm()
	uid,err1 := strconv.Atoi(id)
	if err1!=nil{
		err = err1
		return
	}
	p := Passenger{Id:uid}
	err1 = o.Read(&p)
	if err1 !=nil{
		err = err1
		return
	}
	p.Status = passenger.Status
	_, err2 := o.Update(&p)
	if err2 != nil{
		err = err2
		return
	}
	return &p,nil

}

func DeletePassenger(id string) (err error){
	o := orm.NewOrm()
	uid,err1 := strconv.Atoi(id)
	if err1 != nil{
		err = err1
		return
	}
	p := Passenger{Id:uid}
	_,err2 := o.Delete(&p)
	if err2 != nil{
		err = err2
		return
	}
	return nil
}
