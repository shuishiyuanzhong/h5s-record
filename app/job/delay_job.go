package job

import (
	"encoding/json"
	"fmt"
	customLog "h5s_camera_job/common/log"
	"h5s_camera_job/common/redis"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"sync"
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

type h5sResponse struct {
	BStatus     bool   `json:"bStatus"`
	StrCode     string `json:"strCode"`
	StrFileName string `json:"strFileName"`
	StrUrl      string `json:"strUrl"`
	Record      []struct {
		StrPath string `json:"strPath"`
	} `json:"record"`
}

func (c CameraJob) AddJob() error {
	err := redis.SetCameraJob(c.Id, c.StructToStr())
	if err != nil {
		log.Errorf("err=%v\n", err)
		return err
	}
	return nil
}

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
		return err
	}
	// 先从库中删除旧数据
	err = redis.DeleteCameraJob(strconv.Itoa(c.Id))
	if err != nil {
		return err
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
			return err
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
			return err
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
		return err
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
		response, err := http.Get("http://" + job.IP + "/api/v1/ManualRecordStart?limittime=" +
			strconv.FormatInt(limitTime, 10) + "&token=" + job.Token)
		if err != nil {
			return err
		}
		// 关闭响应
		defer response.Body.Close()

		// 解析响应结果
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}

		ans := new(h5sResponse)
		err = json.Unmarshal(body, ans)
		if err != nil {
			return err
		}

		// 缓存好fileName，以便后续使用
		job.FileName = ans.StrFileName
		err = job.UpdateJob()
		if err != nil {
			return err
		}
		log.Debug("order执行结束")
	} else {
		job := o.GetJob()
		// 获取文件下载路径
		paths, err := getFilePath(job.FileName, job.Token, job.IP)
		if err != nil {
			return err
		}
		// 下载文件
		err = downloadVideo(paths, job.Token, job.IP)
		if err != nil {
			return err
		}
		// 合并mp4
		err = mergeVideo(job.Token, job.FileName)
		if err != nil {
			return err
		}
	}

	return nil
}

// 循环下载远程视频文件，并保存到本地，并生成相应的txt保存文件路径
// @param filepath 文件下载的uri路径
// @param token 摄像机对应的token
// @param ip 摄像机所处在的流媒体服务器地址
func downloadVideo(filepath []string, token string, ip string) error {
	// 创建一个临时文件夹
	err := os.Mkdir(token, 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}
	// 生成一个文件
	recordFileName := token + ".txt"
	recordFile, err := os.OpenFile(token+"/"+recordFileName, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}
	defer recordFile.Close()

	// 并发改造
	// 声明同步工具
	var waitGroup sync.WaitGroup
	// 声明channel,用于限制并发数量
	count := make(chan int, 5)
	// 声明channel，用于存储并发执行时是否出现执行错误问题
	errorSet := make(chan error)

	for i, path := range filepath {
		// 并发调用
		waitGroup.Add(1)
		// 记录并发量
		count <- i
		go func(wg *sync.WaitGroup, i int, downloadPath string) {
			defer wg.Done()
			defer func() {
				<-count
			}()
			// 下载视频文件
			filename := token + "-part" + strconv.Itoa(i) + ".mp4"
			file, err := os.OpenFile(token+"/"+filename, os.O_RDWR|os.O_CREATE, 0755)

			if err != nil && !os.IsExist(err) {
				file.Close()
				errorSet <- err
			}

			//下载文件
			response, err := http.Get("http://" + ip + downloadPath)
			log.Debugf("开始下载录像文件:%v\n", downloadPath)
			defer response.Body.Close()
			if err != nil {
				errorSet <- err
			}

			write, err := io.Copy(file, response.Body)
			if err != nil {
				errorSet <- err
			}
			log.Debugf("成功下载视频文件:%v,共写入%v个字节\n", filename, write)
			if err != nil {
				errorSet <- err
			}

		}(&waitGroup, i, path)

		// 调用协程执行下载操作，将对应的文件名记录到txt中
		filename := token + "-part" + strconv.Itoa(i) + ".mp4"
		_, err = recordFile.Write([]byte("file " + filename + "\n"))
	}
	// 并发调用完毕，等阻塞等待子协程全部执行
	waitGroup.Wait()
	// 全部线程执行完毕，main协程检查并发执行时是否return error
	close(errorSet)
	close(count)
	for err = range errorSet {
		// 出现err，返回
		return err
	}
	return nil
}

// 调用ffmpeg服务，合并视频
func mergeVideo(token string, filename string) error {
	baseDir := "output/"
	err := os.Mkdir(baseDir, 0755)
	if err != nil && !os.IsExist(err) {
		return err
	} else {
		err = nil
	}
	currentPath, err := os.Getwd()
	if err != nil {
		return err
	}
	command := exec.Command("ffmpeg", "-f", "concat", "-i",
		currentPath+"/"+token+"/"+token+".txt", "-c", "copy", baseDir+"/"+filename)
	output, err := command.CombinedOutput()
	if err != nil {
		log.Infof("控制台输出:\n%s\n", string(output))
		fmt.Printf("控制台输出:\n%s\n", string(output))
		return err
	}
	log.Infof("控制台输出:\n%s\n", string(output))
	return nil
}

func getFilePath(filename string, token string, ip string) (ans []string, err error) {
	response, err := http.Get("http://" + ip + "/api/v1/SearchByFilename?type=record&token=" +
		token + "&filename=" + filename)

	if err != nil {
		return nil, err
	}
	// 关闭响应
	defer response.Body.Close()

	// 解析响应结果
	body, err := io.ReadAll(response.Body)
	tmp := new(h5sResponse)

	err = json.Unmarshal(body, tmp)
	if err != nil {
		return nil, err
	}
	for _, s := range tmp.Record {
		ans = append(ans, s.StrPath)
	}
	return
}

func (o OperationOrder) SaveOrder() error {
	err := redis.SetOrder(o.ExecuteTime, o.StructToStr())
	if err != nil {
		return err
	}
	return nil
}

func (o OperationOrder) DeleteOrder() error {
	err := redis.DeleteOrder(o.StructToStr())
	if err != nil {
		log.Errorf("删除Order失败，err=%v", err)
		return err
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
