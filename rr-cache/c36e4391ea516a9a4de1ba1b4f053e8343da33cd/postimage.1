package queue

import (
	"github.com/shuishiyuanzhong/h5s-record/app/job"
	customLog "github.com/shuishiyuanzhong/h5s-record/common/log"
	"time"
)

var log = customLog.Logger()

const ENDING_RECORD_KEY = "h5s:recordSet:Set"

// DelayJob 延时任务，作为用户忘记调用finishRecord接口的保障手段
func DelayJob(payload interface{}) {
	// 将payload转换为CameraJob
	var cameraJob job.CameraJob
	cameraJob = payload.(job.CameraJob)

	// 线程休眠
	currentTime := time.Now().Unix() * 1000
	if currentTime > cameraJob.EndTime {
		// 当前时间大于命令将要执行的时间，终止协程运行
		log.Errorln("执行时间错误，executeTime=%v,currentTime=%v\n", cameraJob.EndTime, currentTime)
		return
	}
	// 休眠
	time.Sleep(time.Duration(cameraJob.EndTime-currentTime) * time.Millisecond)
	err := cameraJob.FinishRecord()
	if err != nil {
		return
	}
}
