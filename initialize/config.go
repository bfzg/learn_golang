package initialize

import (
	"ginchat/global"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func InitConfig() {
	//实例对象
	v := viper.New()

	configFile := "../ginchat/config-debug.yaml"

	v.SetConfigFile(configFile)

	if err := v.ReadConfig(); err != nil {
		panic(err)
	}

	if err := v.Unmarshal(&global.ServiceConfig); err != nil {
		panic(err)
	}

	zap.S().Info("配置信息", global.ServiceConfig)
}
