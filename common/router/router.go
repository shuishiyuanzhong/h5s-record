package router

import (
	"github.com/gin-gonic/gin"
	recordRouter "h5s_camera_job/app/server/record/router"
	customLog "h5s_camera_job/common/log"
)

var router *gin.Engine

func ServerStart() {
	router = gin.New()
	// 使用自定义的日志中间件
	router.Use(customLog.LoggerMiddleware())

	// 加载路由组
	recordRouter.RegisterRouter(router)
	// 监听端口
	err := router.Run(":8888")
	if err != nil {
		return
	}
}
