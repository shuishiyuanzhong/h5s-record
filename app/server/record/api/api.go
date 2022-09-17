package api

import (
	"errors"
	"github.com/gin-gonic/gin"
	"h5s_camera_job/app/job"
	"h5s_camera_job/app/server/record/service"
	"h5s_camera_job/app/utils/payload"
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
	meetingId := c.Param("meetingId")
	if meetingId == "" {
		// 没有传参
		RenderJson(c, errors.New("缺少参数meetingId"), nil)
	}

	err := service.FinishRecord(meetingId)
	RenderJson(c, err, nil)
}

func RenderJson(c *gin.Context, result error, obj interface{}) {
	response := buildPayload(result, obj)
	c.JSON(http.StatusOK, response)
}

func buildPayload(result error, obj interface{}) (payLoad *payload.ResponsePayLoad) {
	payLoad = payload.NewResponsePayLoad(result, obj)
	return
}
