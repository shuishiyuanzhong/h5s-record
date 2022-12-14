package service

import (
	"encoding/json"
	"github.com/shuishiyuanzhong/h5s-record/app/job"
	"github.com/shuishiyuanzhong/h5s-record/app/server/record/service/queue"
	"github.com/shuishiyuanzhong/h5s-record/common/redis"
)

const (
	JOB_KEY = "h5s:job:Job:"
)

func AddJob(job job.CameraJob) error {
	// 将job添加进缓存
	err := job.SaveJob()
	if err != nil {
		return err
	}
	// 开始录像
	err = job.StartRecord()
	if err != nil {
		return err
	}
	// 将job入队，启动结束录像功能(异步执行)
	go queue.DelayJob(job)

	if err != nil {
		return err
	}
	return nil
}

func FinishRecord(meetingId string) error {
	result, err := redis.Get(JOB_KEY + meetingId)
	if err != nil {
		return err
	}
	cameraJob := new(job.CameraJob)
	err = json.Unmarshal([]byte(result.(string)), cameraJob)
	if err != nil {
		return err
	}

	err = cameraJob.FinishRecord()
	if err != nil {
		return err
	}
	return nil
}
