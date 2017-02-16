package controllers

import (
	"didi_api/models"
	"encoding/json"

	"github.com/astaxie/beego"
)

// Operations about Orders
type OrderController struct {
	beego.Controller
}

// @Title CreateOrder
// @Description create orders
// @Param	body		body 	models.Order	true		"body for order content"
// @Success 200 {int} models.Order.Id
// @Failure 403 body is empty
// @router / [post]
func (u *OrderController) Post() {
	var order models.Orders
	json.Unmarshal(u.Ctx.Input.RequestBody, &order)
	uid,err := models.AddOrder(order)
	if err != nil{
		u.Data["json"] = map[string]string{"err":err.Error()}
	}else{
		u.Data["json"] = map[string]int64{"uid": uid}
	}
	u.ServeJSON()
}

// @Title GetAll
// @Description get all Orders
// @Success 200 {object} models.Order
// @router / [get]
func (u *OrderController) GetAll() {
	orders,err := models.GetAllOrders()
	if err != nil{
		u.Data["json"] = map[string]string{"err":err.Error()}
	}else{
		u.Data["json"] = orders
	}
	u.ServeJSON()
}

// @Title Get
// @Description get order by uid
// @Param	uid		path 	string	true		"The key for staticblock"
// @Success 200 {object} models.Order
// @Failure 403 :uid is empty
// @router /:uid [get]
func (u *OrderController) Get() {
	uid := u.GetString(":uid")
	if uid != "" {
		order, err := models.GetOrder(uid)
		if err != nil {
			u.Data["json"] = map[string]string {"err":err.Error()}
		} else {
			u.Data["json"] = order
		}
	}
	u.ServeJSON()
}

// @Title Update
// @Description update the order
// @Param	uid		path 	string	true		"The uid you want to update"
// @Param	body		body 	models.Order	true		"body for user content"
// @Success 200 {object} models.Order
// @Failure 403 :uid is not int
// @router /:uid [put]
func (u *OrderController) Put() {
	uid := u.GetString(":uid")
	if uid != "" {
		var order models.Orders
		json.Unmarshal(u.Ctx.Input.RequestBody, &order)
		d, err := models.UpdateOrder(uid, &order)
		if err != nil {
			u.Data["json"] = map[string]string {"err":err.Error()}
		} else {
			u.Data["json"] = d
		}
	}
	u.ServeJSON()
}

// @Title Delete
// @Description delete the driver
// @Param	uid		path 	string	true		"The uid you want to delete"
// @Success 200 {string} delete success!
// @Failure 403 uid is empty
// @router /:uid [delete]
func (u *OrderController) Delete() {
	uid := u.GetString(":uid")
	err := models.DeleteOrder(uid)
	if err != nil{
		u.Data["json"] = map[string]string{"err":err.Error()}
	}else{
		u.Data["json"] = map[string]string{"success":"delete success!"}
	}
	u.ServeJSON()
}

