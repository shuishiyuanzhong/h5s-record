package controller

import (
	"errors"
	"github.com/gin-gonic/gin"
	//"github.com/juju/errors"
	"github.com/shuishiyuanzhong/h5s-record/app/job"
	"github.com/shuishiyuanzhong/h5s-record/app/server/record/service"
	customLog "github.com/shuishiyuanzhong/h5s-record/common/log"
	"net/http"
)

var logger = customLog.Logger()

func AddJob(c *gin.Context) {
	var cameraJob job.CameraJob
	// 接收参数
	err := c.BindJSON(&cameraJob)
	if err != nil {
		logger.Error(err)
		RenderJson(c, err, nil)
		return
	}
	// 参数校验
	if cameraJob.StartTime > cameraJob.EndTime {
		// 入参有误
		RenderJson(c, err, nil)
		return
	}

	// 调用service中的方法
	err = service.AddJob(cameraJob)
	RenderJson(c, err, nil)
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

	RenderJson(c, err, nil)
}

func FinishRecord(c *gin.Context) {
	meetingId := c.Query("meetingId")
	if meetingId == "" {
		// 没有传参
		RenderJson(c, errors.New("缺少参数meetingId"), nil)
	}

	err := service.FinishRecord(meetingId)
	RenderJson(c, err, nil)
}
