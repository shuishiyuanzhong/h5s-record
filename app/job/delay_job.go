package job

import (
	"encoding/json"
	"github.com/juju/errors"
	"github.com/shuishiyuanzhong/h5s-record/app/utils/video"
	customLog "github.com/shuishiyuanzhong/h5s-record/common/log"
	"github.com/shuishiyuanzhong/h5s-record/common/redis"
	"io"
	"net/http"
	"strconv"
	"time"
)

var log = customLog.Logger()

type CameraJob struct {
	Id        int    `json:"id"`
	Token     string `json:"token"`
	IP        string `json:"ip"`
	StartTime int64  `json:"startTime"`
	EndTime   int64  `json:"endTime"`
	FileName  string `json:"fileName"`
}

type OperationOrder struct {
	Id          int   `json:"id"`
	ExecuteTime int64 `json:"executeTime"`
	Type        int   `json:"type"`
}

const (
	JOB_KEY           = "h5s:job:Job:"
	ENDING_RECORD_KEY = "h5s:recordSet:Set"
)

func (c *CameraJob) SaveJob() error {
	//err := redis.SetCameraJob(c.Id, c.StructToStr())
	// 将job存储到redis中，meetingId作为key
	err := redis.Set(JOB_KEY+strconv.Itoa(c.Id), c)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (c *CameraJob) StartRecord() error {
	// 录制时长
	currentTime := time.Now().Unix() * 1000
	limitTime := (c.EndTime - currentTime) / 1000
	// 调用接口
	response, err := http.Get("http://" + c.IP + "/controller/v1/ManualRecordStart?limittime=" +
		strconv.FormatInt(limitTime, 10) + "&token=" + c.Token)
	if err != nil {
		return errors.Trace(err)
	}
	// 关闭响应
	defer response.Body.Close()

	// 解析响应结果
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return errors.Trace(err)
	}

	ans := new(video.H5sResponse)
	err = json.Unmarshal(body, ans)
	if err != nil {
		return errors.Trace(err)
	}

	// 保存好fileName，以便后续使用
	c.FileName = ans.StrFileName
	// 缓存进redis中
	err = c.SaveJob()
	if err != nil {
		return errors.Trace(err)
	}
	log.Debug("order执行结束")
	return nil
}

func (cameraJob CameraJob) FinishRecord() error {
	// 逻辑修改，查询缓存中如果会议已经有当前会议id的结束记录，如果有 结束执行，如果没有记录，写入记录并执行取消任务
	isMember, err := redis.SIsMember(ENDING_RECORD_KEY, cameraJob.Id)
	if isMember || err != nil {
		// 已经有记录，说明主持人提前取消会议
		return nil
	}
	// 此前没有结束记录，写入新记录
	_, err = redis.SAdd(ENDING_RECORD_KEY, cameraJob.Id)
	if err != nil {
		return errors.Trace(err)
	}

	// 调用服务端接口结束录像
	video.StopRecord(cameraJob.Token, cameraJob.IP)

	// 进行视频文件下载、合并
	err = video.GenerateVideos(cameraJob.FileName, cameraJob.Token, cameraJob.IP)
	if err != nil {
		return errors.Trace(err)
	}
	// 生成视频文件之后，调用后端的服务接口，通知后端下载文件
	err = video.UploadVideoMessage(cameraJob.Id, cameraJob.FileName)
	if err != nil {
		return errors.Trace(err)
	}

	return nil
}

// GetJob 根据会议id获取录像任务的信息
func GetJob(meetingId int) (CameraJob, error) {
	//err := redis.SetCameraJob(c.Id, c.StructToStr())
	// 将job存储到redis中，meetingId作为key
	result, err := redis.Get(JOB_KEY + strconv.Itoa(meetingId))
	if err != nil {
		return CameraJob{}, errors.Trace(err)
	}

	return result.(CameraJob), nil
}

// ParsingOrder Deprecated
func (c *CameraJob) ParsingOrder() []OperationOrder {
	// 抽取出两个指令
	start := OperationOrder{
		Id:          c.Id,
		ExecuteTime: c.StartTime,
		Type:        1,
	}
	end := OperationOrder{
		Id:          c.Id,
		ExecuteTime: c.EndTime,
		Type:        0,
	}
	return []OperationOrder{start, end}
}

func (c *CameraJob) UpdateJob() error {
	// 获得原有的数据
	var old CameraJob
	err := old.StrToStruct(redis.GetCameraJob(c.Id))
	if err != nil {
		return errors.Trace(err)
	}
	// 先从库中删除旧数据
	err = redis.DeleteCameraJob(strconv.Itoa(c.Id))
	if err != nil {
		return errors.Trace(err)
	}

	// 新旧数据比对，如果新数据某字段为空，则将旧数据覆盖到新数据中
	if c.Token == "" {
		c.Token = old.Token
	}
	if c.StartTime == 0 {
		c.StartTime = old.StartTime
	} else if c.StartTime != old.StartTime {
		// job的开始时间有更新，将旧的order删除
		// 删除后publish会重新将order落库
		err = OperationOrder{
			Id:          c.Id,
			ExecuteTime: c.StartTime,
			Type:        1,
		}.DeleteOrder()
		if err != nil {
			return errors.Trace(err)
		}
	}
	if c.EndTime == 0 {
		c.EndTime = old.EndTime
	} else if c.EndTime != old.EndTime {
		// job的结束时间有更新，将旧的order删除
		// 删除后publish会重新将order落库
		err = OperationOrder{
			Id:          c.Id,
			ExecuteTime: c.EndTime,
			Type:        0,
		}.DeleteOrder()
		if err != nil {
			return errors.Trace(err)
		}
	}
	if c.FileName == "" {
		c.FileName = old.FileName
	}
	if c.IP == "" {
		c.IP = old.IP
	}
	// 数据重新进入缓存
	err = redis.SetCameraJob(c.Id, c.StructToStr())
	if err != nil {
		return errors.Trace(err)
	}

	return nil
}

func (o OperationOrder) GetJob() *CameraJob {
	var job CameraJob
	err := job.StrToStruct(redis.GetCameraJob(o.Id))
	if err != nil {
		return nil
	}
	return &job
}

func (o OperationOrder) ExecuteOrder() error {
	log.Debugln("执行命令")
	// type:1为开始录制，0为录制结束
	if o.Type == 1 {
		job := o.GetJob()
		// 录制时长
		limitTime := (job.EndTime - job.StartTime) / 1000
		// 调用接口
		response, err := http.Get("http://" + job.IP + "/controller/v1/ManualRecordStart?limittime=" +
			strconv.FormatInt(limitTime, 10) + "&token=" + job.Token)
		if err != nil {
			return errors.Trace(err)
		}
		// 关闭响应
		defer response.Body.Close()

		// 解析响应结果
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return errors.Trace(err)
		}

		ans := new(video.H5sResponse)
		err = json.Unmarshal(body, ans)
		if err != nil {
			return errors.Trace(err)
		}

		// 缓存好fileName，以便后续使用
		job.FileName = ans.StrFileName
		err = job.UpdateJob()
		if err != nil {
			return errors.Trace(err)
		}
		log.Debug("order执行结束")
	} else {
		job := o.GetJob()
		// 直接调用video中的函数，进行视频文件下载、合并
		err := video.GenerateVideos(job.FileName, job.Token, job.IP)
		if err != nil {
			return errors.Trace(err)
		}
		// 生成视频文件之后，调用后端的服务接口，通知后端下载文件

	}

	return nil
}

func (o OperationOrder) SaveOrder() error {
	err := redis.SetOrder(o.ExecuteTime, o.StructToStr())
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (o OperationOrder) DeleteOrder() error {
	err := redis.DeleteOrder(o.StructToStr())
	if err != nil {
		log.Errorf("删除Order失败，err=%v", err)
		return errors.Trace(err)
	}
	return nil
}

func (o OperationOrder) StructToStr() string {
	return structToStr(o)
}

func (o *OperationOrder) StrToStruct(s string) error {
	err := json.Unmarshal([]byte(s), o)
	if err != nil {
		log.Errorf("StrToStruct失败，err=%v", err)
	}
	return nil
}

func (c CameraJob) StructToStr() string {
	return structToStr(c)
}

func (c *CameraJob) StrToStruct(s string) error {
	err := json.Unmarshal([]byte(s), c)
	if err != nil {
		log.Errorf("StrToStruct失败，err=%v", err)
	}
	return nil
}

// 抽取出来的公共方法，减少代码量
func structToStr(i interface{}) string {
	str, err := json.Marshal(i)
	if err != nil {
		log.Errorf("structToStr失败，err=%v", err)
	}
	return string(str)
}

func (c *CameraJob) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

type Job interface {
	// AddJob 往redis中新增定时任务
	AddJob() error
	// ParsingOrder 根据CameraJob解析出两个Order
	ParsingOrder() []OperationOrder
	// UpdateJob 修改定时任务的数据，在redis中匹配对应的数据，进行删除，然后重新覆盖进去
	UpdateJob() error
}

type Order interface {
	// GetJob 根据order中的id获取Job的信息,当返回nil时说明没有找到相关的Job
	GetJob() *CameraJob
	// ExecuteOrder 执行录像命令
	ExecuteOrder() error
	// SaveOrder 保存Order
	SaveOrder() error
	// DeleteOrder 从redis中删除Order
	DeleteOrder() error
}

type TypeChange interface {
	// StructToStr 类型转换，将结构体转换成字符串
	StructToStr() string
	// StrToStruct 将字符串转换为结构体
	StrToStruct(string) error
}
