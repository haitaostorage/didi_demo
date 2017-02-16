package controllers

import (
	"didi_api/models"
	"encoding/json"

	"github.com/astaxie/beego"
)

// Operations about Drivers
type DriverController struct {
	beego.Controller
}

// @Title CreateDriver
// @Description create drivers
// @Param	body		body 	models.Driver	true		"body for driver content"
// @Success 200 {int} models.Driver.Id
// @Failure 403 body is empty
// @router / [post]
func (u *DriverController) Post() {
	var driver models.Driver
	json.Unmarshal(u.Ctx.Input.RequestBody, &driver)
	uid,err := models.AddDriver(driver)
	if err != nil{
		u.Data["json"] = map[string]string{"err":err.Error()}
	}else{
		u.Data["json"] = map[string]int64{"uid": uid}
	}
	u.ServeJSON()
}
// @Title GetAll
// @Description get all Drivers
// @Success 200 {object} models.Driver
// @router / [get]
func (u *DriverController) GetAll() {
	drivers,err := models.GetAllDrivers()
	if err != nil{
		u.Data["json"] = map[string]string{"err":err.Error()}
	}else{
		u.Data["json"] = drivers
	}
	u.ServeJSON()
}

// @Title Get
// @Description get driver by uid
// @Param	uid		path 	string	true		"The key for staticblock"
// @Success 200 {object} models.Driver
// @Failure 403 :uid is empty
// @router /:uid [get]
func (u *DriverController) Get() {
	uid := u.GetString(":uid")
	if uid != "" {
		driver, err := models.GetDriver(uid)
		if err != nil {
			u.Data["json"] = map[string]string {"err":err.Error()}
		} else {
			u.Data["json"] = driver
		}
	}
	u.ServeJSON()
}

// @Title Update
// @Description update the driver
// @Param	uid		path 	string	true		"The uid you want to update"
// @Param	body		body 	models.User	true		"body for user content"
// @Success 200 {object} models.User
// @Failure 403 :uid is not int
// @router /:uid [put]
func (u *DriverController) Put() {
	uid := u.GetString(":uid")
	if uid != "" {
		var driver models.Driver
		json.Unmarshal(u.Ctx.Input.RequestBody, &driver)
		d, err := models.UpdateDriver(uid, &driver)
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
func (u *DriverController) Delete() {
	uid := u.GetString(":uid")
	err := models.DeleteDriver(uid)
	if err != nil{
		u.Data["json"] = map[string]string{"err":err.Error()}
	}else{
		u.Data["json"] = map[string]string{"success":"delete success!"}
	}
	u.ServeJSON()
}

// @Title charge
// @Description update driver balance
// @Param	uid
// @Success 200 models.Driver
// @Failure 403 uid is empty
// @router /:uid [update]

func(u *DriverController) Charge(){
	uid := u.GetString(":uid")
	var driver models.Driver
	json.Unmarshal(u.Ctx.Input.RequestBody, &driver)
	d,err := models.Charge(uid,&driver)
	if err != nil{
		u.Data["json"] = map[string]string{"err":err.Error()}
	}else{
		u.Data["json"] = d
	}
	u.ServeJSON()
}

