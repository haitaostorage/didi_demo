package controllers

import (
	"didi_api/models"

	"github.com/astaxie/beego"
)

// Operations about order dispatch
type DispatchController struct {
	beego.Controller
}

func (u *DispatchController) Dispatch(){
	uid := u.GetString(":uid")
	if uid != ""{
		driver,err := models.Dispatch(uid)
		if err != nil{
			u.Data["json"] = map[string]string{"err":err.Error()}
		}else{
			u.Data["json"] = driver
		}
	}
	u.ServeJSON()
}


