package initialize

import (
	"log"

	"go.uber.org/zap"
)

func InitLogger() {
	//初始化日志
	logger, err := zap.NewDevelopment()

	if err != nil {
		log.Fatal("初始化日志失败!", err.Error())
	}

	//使用全局 logger
	zap.ReplaceGlobals(logger)
}
