package router

import (
	"github.com/gin-gonic/gin"
	"github.com/shuishiyuanzhong/h5s-record/app/server/record/controller"
)

func RegisterRouter(v1 *gin.Engine) {
	r := v1.Group("/h5s")
	{
		r.POST("/addJob", controller.AddJob)
		//r.PUT("/updateJob", controller.UpdateJob)
		r.GET("/finishJob", controller.FinishRecord)
		//r.DELETE("/deleteJob", controller.DeleteJob)
	}
}
