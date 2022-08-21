package service

import (
	"h5s_camera_job/app/job"
	"h5s_camera_job/app/server/record/service/queue"
	"h5s_camera_job/common/redis"
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
	cameraJob := result.(job.CameraJob)
	err = cameraJob.FinishRecord()
	if err != nil {
		return err
	}
	return nil
}
