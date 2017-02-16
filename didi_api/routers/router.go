// @APIVersion 1.0.0
// @Title beego Test API
// @Description beego has a very cool tools to autogenerate documents for your API
// @Contact astaxie@gmail.com
// @TermsOfServiceUrl http://beego.me/
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html
package routers

import (
	"didi_api/controllers"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"github.com/astaxie/beego"
)

func init() {
	orm.RegisterDriver("mysql", orm.DRMySQL)
	orm.RegisterDataBase("default","mysql","root:123456@tcp(127.0.0.1:3306)/didi_test?charset=utf8",30)
	orm.RunSyncdb("default",false,true)

	ns := beego.NewNamespace("/v1",
		beego.NSNamespace("/passenger",
			beego.NSRouter("/",
				&controllers.PassengerController{},"get:GetAll",
			),
			beego.NSRouter("/?:uid",
				&controllers.PassengerController{},
			),
		),
		beego.NSNamespace("/driver",
			beego.NSRouter("/",
				&controllers.DriverController{},"get:GetAll",
			),
			beego.NSRouter("/?:uid", 
				&controllers.DriverController{},
			),
		),
		beego.NSNamespace("/order",
			beego.NSRouter("/",
				&controllers.OrderController{},"get:GetAll",	
			),
			beego.NSRouter("/?:uid",
				&controllers.OrderController{},
			),	
		),
		beego.NSNamespace("/charge",
			beego.NSRouter("/?:uid",
				&controllers.DriverController{},"put:Charge",
			),
		),
		beego.NSNamespace("/ds",
			beego.NSRouter("/",
				&controllers.DispatchPoolController{},"get:JoinPool",
			),	
			beego.NSRouter("/leave",
				&controllers.DispatchPoolController{},"get:LeavePool",
			),	
		),

	)
	beego.AddNamespace(ns)
}
