package consumer

import (
	"github.com/shuishiyuanzhong/h5s-record/app/job"
	customLog "github.com/shuishiyuanzhong/h5s-record/common/log"
	"github.com/shuishiyuanzhong/h5s-record/common/redis"
	"time"
)

var log = customLog.Logger()

const ENDING_RECORD_KEY = "h5s:recordSet:Set"

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
	currentTime := time.Now().Unix() * 1000
	if currentTime > order.ExecuteTime {
		// 当前时间大于命令将要执行的时间，终止协程运行
		log.Errorln("执行时间错误，executeTime=%v,currentTime=%v\n", order.ExecuteTime, currentTime)
		return
	}

	time.Sleep(time.Duration(order.ExecuteTime-currentTime) * time.Millisecond)

	// 逻辑修改，查询缓存中如果会议已经有当前会议id的结束记录，如果有结束执行，如果没有记录，写入记录并执行取消任务
	isMember, err := redis.SIsMember(ENDING_RECORD_KEY, order.Id)
	if isMember || err != nil {
		// 已经有记录，说明主持人提前取消会议
		return
	}
	// 没有记录，写入新记录
	_, err = redis.SAdd(ENDING_RECORD_KEY, order.Id)
	if err != nil {
		return
	}
	// 执行任务取消录像
	err = order.ExecuteOrder()
	if err != nil {
		return
	}
}
