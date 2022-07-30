package customLog

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"time"
)

var logger = logrus.New()

func Logger() *logrus.Logger {
	return logger
}

func init() {
	// 定制日志输出内容
	logger.Formatter = &logrus.TextFormatter{}
	// 输出文件名
	logger.SetReportCaller(true)

	// 设置日志输出方式
	logfile, _ := os.OpenFile("./logrus.log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	writers := []io.Writer{
		logfile, os.Stdout,
	}
	// 同时输出控制台和文件
	multiWriter := io.MultiWriter(writers...)
	logger.SetOutput(multiWriter)
	// 设置日志等级
	logger.SetLevel(logrus.DebugLevel)
}

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		//开始时间
		startTime := time.Now()
		//处理请求
		c.Next()
		//结束时间
		endTime := time.Now()
		// 执行时间
		latencyTime := endTime.Sub(startTime)
		//请求方式
		reqMethod := c.Request.Method
		//请求路由
		reqUrl := c.Request.RequestURI
		//状态码
		statusCode := c.Writer.Status()
		//请求ip
		clientIP := c.ClientIP()

		// 日志格式
		logger.WithFields(logrus.Fields{
			"status_code":  statusCode,
			"latency_time": latencyTime,
			"client_ip":    clientIP,
			"req_method":   reqMethod,
			"req_uri":      reqUrl,
		}).Info()
	}
}
