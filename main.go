package main

import (
	"log"
	"todo-api/models/db"
	_ "todo-api/routers"

	beego "github.com/beego/beego/v2/server/web"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	db.InitDB()
	beego.Run()
}

