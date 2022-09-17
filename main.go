package main

import (
	"github.com/shuishiyuanzhong/h5s_record/app/consumer"
	"github.com/shuishiyuanzhong/h5s_record/common/router"
)

func main() {
	// 启动消费者
	go consumer.ReceiveOrder()

	router.ServerStart()
}
