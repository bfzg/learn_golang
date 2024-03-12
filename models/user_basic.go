package models

import (
	"time"

	"gorm.io/gorm"
)

type Model struct {
	ID        uint      `gorm:"primaryKey"` //这里第三个选项是数据标签 用来指定数据来源
	CreatedAt time.Time // 注意可以被外部访问到的数据 是大写，私有小写
	UpdateAt  time.Time
	DeleteAt  gorm.DeletedAt `gorm:"index"`
}

type UserBasic struct {
	Model
	Name     string
	PassWord string
	Avatar   string
	//下面的表示在gorm 中 column:gender：指定该字段在数据库表中的列名为"gender"。
	//default:male：指定该字段在数据库中的默认值为"male"。
	//type:varchar(6)：指定该字段在数据库中的类型为varchar，长度为6个字符。
	//comment 'male表示男, famale表示女'：添加注释，描述该字段的含义，便于理解和维护。
	Gender        string `gorm:"column:gender;default:male;type:varchar(6) comment 'male表示男, famale表示女'"`
	Phone         string `valid:"matches(^1[3-9]{1}\\d{9}$)"` //valid为条件约束
	Email         string `valid:"email"`
	Identity      string
	ClientIp      string `valid:"ipv4"`
	ClientPort    string
	Salt          string     //盐值
	LoginTime     *time.Time `gorm:"column:login_time"`
	HeartBeatTime *time.Time `gorm:"column:heart_beat_time"`
	LoginOutTime  *time.Time `gorm:"column:login_out_time"`
	IsLoginOut    bool
	DeviceInfo    string //登录设备
}

func (table *UserBasic) UserTableName() string {
	return "user_basic"
}
