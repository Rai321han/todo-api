package routers

import (
	"todo-api/controllers"

	beego "github.com/beego/beego/v2/server/web"
)

func init() {
	ns := beego.NewNamespace("/v1/",
		beego.NSNamespace("/api",
			beego.NSNamespace("/auth",
				beego.NSRouter("/register", &controllers.AuthController{}, "post:Register"),
				beego.NSRouter("/login", &controllers.AuthController{}, "post:Login"),
				beego.NSRouter("/logout", &controllers.AuthController{}, "post:Logout"),
				beego.NSRouter("/me", &controllers.AuthController{}, "get:Me"),
			),

			beego.NSNamespace("/todos",
				beego.NSRouter("/", &controllers.TodoController{}, "get:GetAll;post:Create"),
				beego.NSRouter("/:id", &controllers.TodoController{}, "get:GetByID;put:Update;delete:Delete"),
			),
		),
	)

	beego.AddNamespace(ns)
}
