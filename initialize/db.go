package initialize

import (
	"fmt"
	"ginchat/global"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 初始化数据库
func InitDB() {
	//这里 %s 输出字符串  %d 整数十进制
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", "root", "root", "127.0.0.1", 3306, "ginchat")

	//写sql 配置
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second, // 慢 SQL 阈值
			LogLevel:                  logger.Info, //日志级别
			IgnoreRecordNotFoundError: true,        //忽略ErrRecordNotFound（记录未找到）错误
			Colorful:                  true,        //禁用彩色打印
		},
	)

	var err error

	global.DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger, //打印sql日志
	})

	if err != nil {
		panic(err)
	}
}

func InitRedis() {
	opt := redis.Options{
		Addr:     fmt.Sprintf("%s:%d", global.ServiceConfig.RedisDB.Host, global.ServiceConfig.RedisDB.Port), // redis地址
		Password: "",                                                                                         // redis密码，没有则留空
		DB:       10,                                                                                         // 默认数据库，默认是0
	}
	global.RedisDB = redis.NewClient(&opt)
}
