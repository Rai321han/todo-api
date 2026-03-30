package main

import (
	"todo-api/middlewares"
	"todo-api/models/db"
	_ "todo-api/routers"

	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/filter/cors"
)

func main() {
	frontendURL, _ := beego.AppConfig.String("frontendurl")
	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowOrigins:     []string{frontendURL},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	beego.BConfig.RecoverPanic = false
	beego.InsertFilterChain("*", middlewares.RecoveryFilterChain)
	beego.InsertFilterChain("*", middlewares.RequestLogger)
	db.InitDB()
	beego.Run()
}
