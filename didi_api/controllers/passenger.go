package controllers

import (
	"didi_api/models"
	"encoding/json"
	"github.com/astaxie/beego"
)

// Operations about Passengers
type PassengerController struct {
	beego.Controller
}

// @Title CreatePassenger
// @Description create passenger
// @Param	body		body 	models.Passenger	true	"body for passenger content"
// @Success 200 {int} models.Passenger.Id
// @Failure 403 body is empty
// @router / [post]
func (u *PassengerController) Post() {
	var passenger models.Passenger
	json.Unmarshal(u.Ctx.Input.RequestBody, &passenger)
	if passenger.Name == ""{
		u.Data["json"] = map[string]string{"err":"Name should not be empty"};
		u.ServeJSON()
	}else{
		uid,err := models.AddPassenger(passenger)
		if err != nil{
			u.Data["json"] = map[string]string{"err":err.Error()}
		}else{
			u.Data["json"] = map[string]int64{"id": uid}
		}
		u.ServeJSON()
	}
}
// @Title GetAll
// @Description get all Passengers
// @Success 200 {object} models.Passenger
// @router / [get]
func (u *PassengerController) GetAll() {
	passengers,err := models.GetAllPassengers()
	if err != nil{
		u.Data["json"] = map[string]string{"err":err.Error()}
	}else{
		u.Data["json"] = passengers
	}
	u.ServeJSON()
}

// @Title Get
// @Description get passenger by uid
// @Param	uid		path 	string	true		"The key for staticblock"
// @Success 200 {object} models.Passenger
// @Failure 403 :uid is empty
// @router /:uid [get]
func (u *PassengerController) Get() {
	uid := u.GetString(":uid")
	if uid != "" {
		passenger, err := models.GetPassenger(uid)
		if err != nil {
			u.Data["json"] = map[string] string{"err":err.Error()}
		} else {
			u.Data["json"] = passenger
		}
	}
	u.ServeJSON()
}

// @Title Update
// @Description update the passenger
// @Param	uid		path 	string	true		"The uid you want to update"
// @Param	body		body 	models.Passenger	true		"body for passenger content"
// @Success 200 {object} models.Passenger
// @Failure 403 :uid is not int
// @router /:uid [put]
func (u *PassengerController) Put() {
	uid := u.GetString(":uid")
	if uid != "" {
		var passenger models.Passenger
		json.Unmarshal(u.Ctx.Input.RequestBody, &passenger)
		uu, err := models.UpdatePassenger(uid, &passenger)
		if err != nil {
			u.Data["json"] = map[string]string{"err":err.Error()}
		} else {
			u.Data["json"] = uu
		}
	}
	u.ServeJSON()
}

// @Title Delete
// @Description delete the passenger
// @Param	uid		path 	string	true		"The uid you want to delete"
// @Success 200 {string} delete success!
// @Failure 403 uid is empty
// @router /:uid [delete]
func (u *PassengerController) Delete() {
	uid := u.GetString(":uid")
	err := models.DeletePassenger(uid)
	if err != nil{
		u.Data["json"] = map[string]string{"err":err.Error()}
	}else{
		u.Data["json"] = map[string]string{"success":"delete success!"}
	}
	u.ServeJSON()
}

