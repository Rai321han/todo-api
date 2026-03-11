package main

import (
	"todo-api/middlewares"
	"todo-api/models/db"
	_ "todo-api/routers"

	beego "github.com/beego/beego/v2/server/web"
)

func main() {
	beego.InsertFilter("*", beego.BeforeRouter, middlewares.RecoveryFilter)
	db.InitDB()
	beego.Run()
}

