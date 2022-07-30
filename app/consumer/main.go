package consumer

import (
	"h5s_camera_job/app/job"
	customLog "h5s_camera_job/common/log"
	"h5s_camera_job/common/redis"
	"runtime"
	"time"
)

var log = customLog.Logger()

// ReceiveOrder 消费者主线程，持续监听channel，
// 接收到数据则创建一个新的协程来执行Order
func ReceiveOrder() {
	rdb := redis.Rdb()
	ctx := *redis.Ctx()

	pubSub := rdb.Subscribe(ctx, "delay_queue")

	defer pubSub.Close()

	// 将go中的channel和redis的绑定起来
	c := pubSub.Channel()
	// 循环监听消息
	log.Debugln("start listen")
	for message := range c {
		// 成功获取到数据，创建一个线程执行
		log.Infof("接收到message，message=%v\n", message.Payload)
		go delayJob(message.Payload)
	}

}

// 延迟任务
func delayJob(payload string) {
	// 将payload转换为Order
	var order job.OperationOrder
	err := order.StrToStruct(payload)
	if err != nil {
		log.Errorln(err)
		return
	}
	// 线程休眠
	currentTime := time.Now().UnixMilli()
	if currentTime > order.ExecuteTime {
		// 当前时间大于命令将要执行的时间，终止协程运行
		log.Errorln("执行时间错误，executeTime=%v,currentTime=%v\n", order.ExecuteTime, currentTime)
		return
	}

	time.Sleep(time.Duration(order.ExecuteTime-currentTime) * time.Millisecond)

	// 唤醒后，从zset中查询order
	err = order.DeleteOrder()
	if err != nil {
		log.Printf("获取order失败，order不存在，协程运行终止")
		// 终止协程
		runtime.Goexit()
	}
	// 执行命令
	err = order.ExecuteOrder()
	if err != nil {
		return
	}
}
