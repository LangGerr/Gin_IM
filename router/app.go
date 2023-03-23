package router

import (
	"GinChat/docs"
	"GinChat/service"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Router() *gin.Engine {
	r := gin.Default()
	// swagger
	//r.Use(middleware.CORSMiddleware())
	docs.SwaggerInfo.BasePath = ""
	// 这一步应该是使用swagger中的页面 引入gin-swagger渲染文档数据
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.Static("/asset", "asset/")
	r.LoadHTMLGlob("views/**/*")

	// 首页
	r.GET("/", service.GetIndex)
	r.GET("/index", service.GetIndex)
	r.GET("/toRegister", service.ToRegister)
	r.GET("/toChat", service.ToChat)
	r.GET("/chat", service.Chat)
	r.POST("/searchFriends", service.SearchFriends)

	// 用户
	r.POST("/user/getUserList", service.GetUserList)
	r.POST("/user/createUser", service.CreateUser)
	r.POST("/user/deleteUser", service.DeleteUser)
	r.POST("/user/updateUser", service.UpdateUser)
	r.POST("/user/findUserByNameAndPwd", service.FindUserByNameAndPwd)
	// 发送消息
	r.GET("/user/sendMsg", service.SendMsg)
	r.GET("/user/sendUserMsg", service.SendUserMsg)
	// 添加好友
	r.POST("/attach/upload", service.Upload)
	// 上传文件
	r.POST("/contact/addfriend", service.AddFriends)
	// 创建群
	r.POST("/contact/createCommunity", service.CreateCommunity)
	// 群列表
	r.POST("/contact/loadcommunity", service.LoadCommunity)
	r.POST("/contact/joinGroup", service.JoinGroups)
	return r
}
