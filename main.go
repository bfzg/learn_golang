package main

import "ginchat/initialize"

func main() {

	//初始化日志
	initialize.InitLogger()

	//初始化数据库
	initialize.InitDB()
}
