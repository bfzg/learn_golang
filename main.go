package main

import (
	"ginchat/initialize"
	"ginchat/router"
)

func main() {

	//初始化日志
	initialize.InitLogger()

	//初始化配置
	initialize.InitConfig()

	//初始化数据库
	initialize.InitDB()
	// initialize.In

	router := router.Router()
	router.Run(":8000")
}
