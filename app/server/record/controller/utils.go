package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/juju/errors"
	"github.com/shuishiyuanzhong/h5s-record/app/utils/payload"
	"net/http"
)

func RenderJson(c *gin.Context, err error, obj interface{}) {
	response := buildPayload(err, obj)
	if err != nil {
		logger.Error(err)
		logger.Error(errors.ErrorStack(err))
		return
	}
	c.JSON(http.StatusOK, response)
}

func buildPayload(err error, obj interface{}) (payLoad *payload.ResponsePayLoad) {
	payLoad = payload.NewResponsePayLoad(err, obj)
	return
}
