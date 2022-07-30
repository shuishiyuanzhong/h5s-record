package main

import (
	"h5s_camera_job/app/consumer"
	"h5s_camera_job/common/router"
)

func main() {
	// 启动消费者
	go consumer.ReceiveOrder()

	router.ServerStart()
}
