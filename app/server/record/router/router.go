package router

import (
	"github.com/gin-gonic/gin"
	"github.com/shuishiyuanzhong/h5s-record/app/server/record/api"
)

func RegisterRouter(v1 *gin.Engine) {
	r := v1.Group("/h5s")
	{
		r.POST("/addJob", api.AddJob)
		//r.PUT("/updateJob", api.UpdateJob)
		r.GET("/finishJob", api.FinishRecord)
		//r.DELETE("/deleteJob", api.DeleteJob)
	}
}
