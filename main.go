package main

import (
	"todo-api/middlewares"
	"todo-api/models/db"
	_ "todo-api/routers"

	beego "github.com/beego/beego/v2/server/web"
)

func main() {
    beego.BConfig.RecoverPanic = false
    beego.InsertFilterChain("*", middlewares.RecoveryFilterChain)
	beego.InsertFilterChain("*", middlewares.RequestLogger)
    db.InitDB()
    beego.Run()
}