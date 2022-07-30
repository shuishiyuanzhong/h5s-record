package service

import (
	"h5s_camera_job/app/job"
	"h5s_camera_job/app/producer"
)

func AddJob(job job.CameraJob) error {
	// 添加进缓存
	err := job.AddJob()
	if err != nil {
		return err
	}
	// job进入缓存后，调用producer，解析出order，同时加入缓存和publish
	err = producer.PubMessage(job)
	if err != nil {
		return err
	}
	return nil
}

func UpdateJob(cameraJob job.CameraJob) error {
	err := cameraJob.UpdateJob()
	if err != nil {
		return err
	}
	// 修改之后重新发放指令
	err = producer.PubMessage(cameraJob)
	if err != nil {
		return err
	}
	return nil
}
