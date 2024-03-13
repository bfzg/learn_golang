package main

import (
	"ginchat/initialize"
	"ginchat/router"
)

func main() {

	//初始化日志
	initialize.InitLogger()

	//初始化数据库
	initialize.InitDB()

	router := router.Router()
	router.Run(":8000")
}
