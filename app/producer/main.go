package producer

import (
	"h5s_camera_job/app/job"
	customLog "h5s_camera_job/common/log"
	"h5s_camera_job/common/redis"
)

var logger = customLog.Logger()

func PubMessage(job job.CameraJob) error {
	rdb := redis.Rdb()
	ctx := *redis.Ctx()
	// 根据job解析出两个指令，同时将指令推送到channel
	for _, order := range job.ParsingOrder() {
		err := order.SaveOrder()
		if err != nil {
			return err
		}
		logger.Infof("publish message:%v\n", order.StructToStr())
		err = rdb.Publish(ctx, "delay_queue", order.StructToStr()).Err()
		if err != nil {
			logger.Errorln(err)
		}
	}
	return nil
}
