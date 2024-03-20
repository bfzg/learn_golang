package router

import (
	"ginchat/middlewear"
	"ginchat/service"

	"github.com/gin-gonic/gin"
)

func Router() *gin.Engine {

	router := gin.Default()
	//v1版本
	v1 := router.Group("v1")

	//用户模块，后续有个用户的api就放置其中
	user := v1.Group("user")
	{
		user.GET("/list", middlewear.JWY(), service.List)
		user.POST("/login_pw", middlewear.JWY(), service.LoginByNameAndPassWord)
		user.POST("/new", middlewear.JWY(), service.NewUser)
		user.DELETE("/delete", middlewear.JWY(), service.DeleteUser)
		user.POST("/updata", middlewear.JWY(), service.UpdataUser)
		user.GET("/SendUserMsg", middlewear.JWY(), service.SendUserMsg)
	}

	//关系
	relation := v1.Group("relation").Use(middlewear.JWY())
	{
		relation.POST("/list", service.FriendList)
		relation.POST("/add", service.AddFriendByName)
		relation.POST("/new_group", service.NewGroup)
		relation.POST("/group_list", service.GroupList)
		relation.POST("/join_group", service.JoinGroup)
	}

	upload := v1.Group("upload").Use(middlewear.JWY())
	{
		upload.POST("/image", service.Image)
	}

	return router
}
