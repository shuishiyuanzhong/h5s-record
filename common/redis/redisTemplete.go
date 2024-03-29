package redis

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/juju/errors"
	customLog "github.com/shuishiyuanzhong/h5s-record/common/log"
	"strconv"
)

var rdb *redis.Client
var ctx context.Context
var (
	jobKey   = "h5s:jobList:Map"
	orderKey = "h5s:delayQueue:ZSet"
	log      = customLog.Logger()
)

func Rdb() *redis.Client {
	return rdb
}

func Ctx() *context.Context {
	return &ctx
}

func init() {
	// 建立连接
	rdb = redis.NewClient(&redis.Options{
		Addr:     "r-wz9o4bhpwhhknbopnfpd.redis.rds.aliyuncs.com:6379",
		Password: "kouse:Gu000803", // no password set
		DB:       0,                // use default DB
	})

	// 获取跟踪上下文
	ctx = context.Background()
}

// GetCameraJob 根据id获取Map中的数据
func GetCameraJob(id int) string {
	result, err := rdb.HGet(ctx, jobKey, strconv.Itoa(id)).Result()
	if err != nil {
		log.Errorf("获取CameraJob失败，id=%v，err=%v\n", id, err)
	}
	// 字符串转义struct
	return result
}

// SetCameraJob 新增任务
func SetCameraJob(id int, member string) error {
	result, err := rdb.HSet(ctx, jobKey, id, member).Result()
	if err != nil {
		log.Errorf("存储CameraJob失败，id=%v，err=%v\n", id, err)
	}
	if result == 0 {
		return errors.New("添加Job不成功")
	}
	return nil
}

// DeleteCameraJob 删除任务
func DeleteCameraJob(id string) error {
	result, err := rdb.HDel(ctx, jobKey, id).Result()
	if err != nil {
		log.Errorf("删除CameraJob失败，id=%v，err=%v\n", id, err)
		return errors.Trace(err)
	}
	if result == 0 {
		return errors.New("删除记录数量为0")
	}
	return nil
}

// SetOrder 新增操作指令
func SetOrder(executeTime int64, member string) error {
	result, err := rdb.ZAdd(ctx, orderKey, &redis.Z{
		Score:  float64(executeTime),
		Member: member,
	}).Result()
	if err != nil {
		log.Errorf("存储OperationOder失败，err=%v\n", err)
		return errors.Trace(err)
	}
	if result == 0 {
		return errors.New("新增记录数量为0")
	}
	log.Debugf("存储OperationOder成功，order=%v\n", member)
	return nil
}

// DeleteOrder 删除特定的操作命令
func DeleteOrder(order string) error {
	result, err := rdb.ZRem(ctx, orderKey, order).Result()
	if err != nil {
		log.Errorf("删除OperationOder失败，err=%v\n", err)
	}
	if result == 0 {
		return errors.New("删除记录数量为0")
	}
	log.Debugf("删除OperationOder成功，order=%v\n", order)
	return nil
}

func SAdd(key string, value interface{}) (bool, error) {
	result, err := rdb.SAdd(ctx, key, value).Result()
	if err != nil {
		return false, errors.Trace(err)
	}
	return result > 0, nil
}

func SIsMember(key string, value interface{}) (bool, error) {
	result, err := rdb.SIsMember(ctx, key, value).Result()
	if err != nil {
		return false, errors.Trace(err)
	}
	return result, nil
}

// Get redis get接口
func Get(key string) (interface{}, error) {
	result, err := rdb.Get(ctx, key).Result()
	if err != nil {
		return nil, errors.Trace(err)
	}
	return result, nil
}

// Set redis set接口
func Set(key string, value interface{}) error {
	_, err := rdb.Set(ctx, key, value, 0).Result()
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}
