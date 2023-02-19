package video

import (
	"encoding/json"
	"fmt"
	"github.com/juju/errors"
	customLog "github.com/shuishiyuanzhong/h5s-record/common/log"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

var log = customLog.Logger()

func GenerateVideos(filename, token, ip string) error {
	// 获取文件下载路径
	paths, err := getFilePath(filename, token, ip)
	if err != nil {
		return errors.Trace(err)
	}
	// 下载文件
	err = downloadVideo(paths, token, ip)
	if err != nil {
		return errors.Trace(err)
	}
	// 合并mp4
	err = mergeVideo(token, filename)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

// DownloadVideo 循环下载远程视频文件，并保存到本地，并生成相应的txt保存文件路径
// @param filepath 文件下载的uri路径
// @param token 摄像机对应的token
// @param ip 摄像机所处在的流媒体服务器地址
func downloadVideo(filepath []string, token string, ip string) error {
	// 创建一个临时文件夹
	err := os.Mkdir(token, 0755)
	if err != nil && !os.IsExist(err) {
		return errors.Trace(err)
	}
	// 生成一个文件
	recordFileName := token + ".txt"
	recordFile, err := os.OpenFile(token+"/"+recordFileName, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil && !os.IsExist(err) {
		return errors.Trace(err)
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
				errorSet <- errors.Trace(err)
			}

			//下载文件
			response, err := http.Get("http://" + ip + downloadPath)
			log.Debugf("开始下载录像文件:%v\n", downloadPath)
			defer response.Body.Close()
			if err != nil {
				errorSet <- errors.Trace(err)
			}

			write, err := io.Copy(file, response.Body)
			if err != nil {
				errorSet <- errors.Trace(err)
			}
			log.Debugf("成功下载视频文件:%v,共写入%v个字节\n", filename, write)
			if err != nil {
				errorSet <- errors.Trace(err)
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
		return errors.Trace(err)
	}
	return nil
}

// MergeVideo 调用ffmpeg服务，合并视频
func mergeVideo(token string, filename string) error {
	baseDir := "output/"
	err := os.Mkdir(baseDir, 0755)
	if err != nil && !os.IsExist(err) {
		return errors.Trace(err)
	} else {
		err = nil
	}
	currentPath, err := os.Getwd()
	if err != nil {
		return errors.Trace(err)
	}
	command := exec.Command("ffmpeg", "-f", "concat", "-i",
		currentPath+"\\"+token+"\\"+token+".txt", "-c", "copy", baseDir+"/"+filename)
	output, err := command.CombinedOutput()
	if err != nil {
		log.Infof("控制台输出:\n%s\n", string(output))
		fmt.Printf("控制台输出:\n%s\n", string(output))
		return errors.Trace(err)
	}
	log.Infof("控制台输出:\n%s\n", string(output))
	// remove file
	err = os.RemoveAll(token)
	if err != nil {
		fmt.Println(err)
	}
	return nil
}

func getFilePath(filename string, token string, ip string) (ans []string, err error) {
	response, err := http.Get("http://" + ip + "/controller/v1/SearchByFilename?type=record&token=" +
		token + "&filename=" + filename)

	if err != nil {
		return nil, errors.Trace(err)
	}
	// 关闭响应
	defer response.Body.Close()

	// 解析响应结果
	body, err := io.ReadAll(response.Body)
	tmp := new(H5sResponse)

	err = json.Unmarshal(body, tmp)
	if err != nil {
		return nil, errors.Trace(err)
	}
	for _, s := range tmp.Record {
		ans = append(ans, s.StrPath)
	}
	return
}

// UploadVideoMessage 上传视频信息至中控后台
func UploadVideoMessage(meetingId int, filename string) error {
	var request = &ControlRequest{
		MeetingId: meetingId,
		VideoUrl:  "http://8.129.32.153/video/" + filename,
	}

	bytes, err := json.Marshal(request)
	if err != nil {
		return errors.Trace(err)
	}

	response, err := http.Post("http://47.115.32.14:8082/screenControl/platform/meeting/saveVideoRecord",
		"application/json", strings.NewReader(string(bytes)))
	// 关闭响应
	defer response.Body.Close()

	// TODO 解析响应结果
	result, err := io.ReadAll(response.Body)
	if err != nil {
		return errors.Trace(err)
	}
	fmt.Println("request body" + string(bytes))
	fmt.Println("中控response：" + string(result))
	resp := new(centerResponse)
	err = json.Unmarshal(result, resp)
	if err != nil {
		return errors.Trace(err)
	}
	if resp.Code != 200 {
		return fmt.Errorf("err_msg=%v", resp.Msg)
	}

	return nil
}

func StopRecord(token, ip string) {
	// 调用接口
	http.Get("http://" + ip + "/controller/v1/ManualRecordStop?token=" + token)
}

type ControlRequest struct {
	MeetingId int    `json:"meetingId"`
	VideoUrl  string `json:"videoUrl"`
}

type centerResponse struct {
	Code       int         `json:"code"`
	Data       interface{} `json:"data"`
	ExtendData interface{} `json:"extendData"`
	Msg        string      `json:"msg"`
}
