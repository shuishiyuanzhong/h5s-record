package api

import (
	"github.com/gin-gonic/gin"
	"h5s_camera_job/app/job"
	"h5s_camera_job/app/server/record/service"
	customLog "h5s_camera_job/common/log"
	"net/http"
)

var logger = customLog.Logger()

func AddJob(c *gin.Context) {
	var cameraJob job.CameraJob
	// 接收参数
	err := c.BindJSON(&cameraJob)
	if err != nil {
		logger.Error(err)
		c.JSONP(http.StatusBadRequest, err)
		return
	}
	// 参数校验
	if cameraJob.StartTime > cameraJob.EndTime {
		// 入参有误
		c.JSONP(http.StatusBadRequest, "时间有误")
		return
	}

	// 调用service中的方法
	err = service.AddJob(cameraJob)
	if err != nil {
		c.JSONP(http.StatusInternalServerError, err)
		return
	}
	// 响应
	c.JSONP(http.StatusOK, "执行成功")
}

func UpdateJob(c *gin.Context) {
	var cameraJob job.CameraJob
	// 接收参数
	err := c.BindJSON(&cameraJob)
	if err != nil {
		logger.Error(err)
		c.JSONP(http.StatusBadRequest, err)
		return
	}
	// 参数校验
	if cameraJob.StartTime > cameraJob.EndTime {
		// 入参有误
		c.JSONP(http.StatusBadRequest, "时间有误")
		return
	}
	logger.Debugln(cameraJob)

	err = service.UpdateJob(cameraJob)
	if err != nil {
		logger.Debugln(err)
		c.JSONP(http.StatusInternalServerError, err)
		return
	}
	// 响应
	c.JSONP(http.StatusOK, "执行成功")
}

func DeleteJob(c *gin.Context) {

}
