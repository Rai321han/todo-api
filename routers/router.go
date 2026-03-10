package routers

import (
	"todo-api/controllers"
	"todo-api/middlewares"

	beego "github.com/beego/beego/v2/server/web"
)

func init() {
	ns := beego.NewNamespace("/v1/",
		beego.NSNamespace("/api",
				beego.NSNamespace("/auth",
					beego.NSRouter("/register", &controllers.AuthController{}, "post:Register"),
					beego.NSRouter("/login", &controllers.AuthController{}, "post:Login"),
				),
	
				beego.NSNamespace("/todos",
					beego.NSBefore(middlewares.AuthMiddleware),
					beego.NSRouter("/", &controllers.TodoController{}, "post:Create"),
					beego.NSRouter("/:id", &controllers.TodoController{}, "get:GetByID;put:Update;delete:Delete"),
					beego.NSRouter("/all", &controllers.TodoController{}, "get:GetAll"),
				),
		),
		
	)

	beego.AddNamespace(ns)
}
