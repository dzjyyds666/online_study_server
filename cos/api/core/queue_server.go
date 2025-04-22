package core

import (
	"bufio"
	"context"
	"cos/api/config"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/dzjyyds666/opensource/common"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/redis/go-redis/v9"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"sync/atomic"
	"time"
)

const (
	MAX_COROUTINE_NUMBER = 5
)

type LambdaQueueServer struct {
	ctx             context.Context
	hcli            *http.Client
	cosServ         *CosFileServer
	coroutineNumber int64 // 协程的数量
	bucket          string
	tmpDir          string
}

func NewLambdaQueueServer(ctx context.Context, hcli *http.Client, server *CosFileServer) *LambdaQueueServer {
	return &LambdaQueueServer{
		ctx:     ctx,
		hcli:    hcli,
		cosServ: server,
		bucket:  aws.ToString(config.GloableConfig.S3.Bucket[0]),
		tmpDir:  aws.ToString(config.GloableConfig.TmpDir),
	}
}

func (lqs *LambdaQueueServer) Start(ctx context.Context) {
	//err := lqs.WatchQueue(ctx, buildVideoLambdaQueueKey())
	//if err != nil {
	//	logx.GetLogger("study").Errorf("LambdaQueueServer|Start|WatchQueue Error|%v", err)
	//	return
	//}
}

func (lqs *LambdaQueueServer) WatchQueue(ctx context.Context, queueName string) error {
	// 开始监听队列
	logx.GetLogger("study").Infof("LambdaQueueServer|WatchQueue|Start|%s", queueName)
	for {
		if lqs.coroutineNumber >= MAX_COROUTINE_NUMBER {
			// 超过最大异步数量，暂时不处理
			// 延迟1秒
			time.Sleep(1 * time.Second)
			continue
		}
		result, err := lqs.cosServ.cosDB.BRPop(ctx, 5*time.Second, queueName).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				continue
			} else {
				panic(err)
			}
		}
		if len(result) == 0 {
			logx.GetLogger("study").Errorf("LambdaQueueServer|WatchQueue|BRPop Error|%v", err)
			continue
		}
		atomic.AddInt64(&lqs.coroutineNumber, 1)
		//执行视频的处理
		err = lqs.FormatVideo(ctx, result[1])
		if err != nil {
			logx.GetLogger("study").Errorf("LambdaQueueServer|WatchQueue|FormatVideo Error|%v|%s", err, common.ToStringWithoutError(result))
		}
	}
}

func (lqs *LambdaQueueServer) FormatVideo(ctx context.Context, fid string) error {
	//先查询文件的信息
	file, err := lqs.cosServ.QueryCosFile(ctx, fid)
	if err != nil {
		logx.GetLogger("study").Errorf("LambdaQueueServer|FormatVideo|QueryCosFile Error|%v|%s", err, common.ToStringWithoutError(file))
		return err
	}
	// 从客户端中拉取下来文件
	r, err := lqs.cosServ.GetFile(ctx, lqs.bucket, file)
	if err != nil {
		logx.GetLogger("study").Errorf("LambdaQueueServer|FormatVideo|GetFile Error|%v|%s", err, common.ToStringWithoutError(file))
		return err
	}
	defer r.Close()
	// 把文件写入到临时路径
	originPath := path.Join(lqs.tmpDir, aws.ToString(file.Fid)+path.Ext(aws.ToString(file.FileName)))
	err = lqs.SaveFile(ctx, file, originPath, r)
	if err != nil {
		logx.GetLogger("study").Errorf("LambdaQueueServer|FormatVideo|SaveFile Error|%v|%s", err, common.ToStringWithoutError(file))
		return err
	}

	params := []string{
		"-i", originPath,
		"-filter_complex",
		"[0:v]split=3[v360][v720][v1080];" +
			"[v360]scale=w=640:h=360[vv360];" +
			"[v720]scale=w=1280:h=720[vv720];" +
			"[v1080]scale=w=1920:h=1080[vv1080]",
		"-map", "[vv360]", "-map", "a",
		"-c:v:0", "libx264", "-b:v:0", "800k",
		"-c:a:0", "aac", "-ac", "2", "-b:a:0", "96k",
		"-f", "hls", "-hls_time", "6", "-hls_playlist_type", "vod",
		"-hls_segment_filename", "output/360p/segment_%03d.ts", "output/360p/index.m3u8",

		"-map", "[vv720]", "-map", "a",
		"-c:v:1", "libx264", "-b:v:1", "2000k",
		"-c:a:1", "aac", "-ac", "2", "-b:a:1", "128k",
		"-f", "hls", "-hls_time", "6", "-hls_playlist_type", "vod",
		"-hls_segment_filename", "output/720p/segment_%03d.ts", "output/720p/index.m3u8",

		"-map", "[vv1080]", "-map", "a",
		"-c:v:2", "libx264", "-b:v:2", "5000k",
		"-c:a:2", "aac", "-ac", "2", "-b:a:2", "192k",
		"-f", "hls", "-hls_time", "6", "-hls_playlist_type", "vod",
		"-hls_segment_filename", "output/1080p/segment_%03d.ts", "output/1080p/index.m3u8",
	}

	// 把文件转换为hls格式
	err = Exec(ctx, "D:\\utils\\ffmpeg-7.0.2-full_build-shared\\bin\\ffmpeg.exe", params...)
	if err != nil {
		logx.GetLogger("study").Errorf("LambdaQueueServer|FormatVideo|Exec Error|%v|%s", err, common.ToStringWithoutError(file))
		return err
	}
	return nil
}

func (lqs *LambdaQueueServer) SaveFile(ctx context.Context, file *CosFile, targetPath string, reader io.Reader) error {
	openFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		logx.GetLogger("study").Errorf("LambdaQueueServer|SaveFile|OpenFile Error|%v|%s", err, common.ToStringWithoutError(file))
		return err
	}
	buf := make([]byte, 1024*1024*5) // 5M
	for {
		n, err := reader.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			logx.GetLogger("study").Errorf("LambdaQueueServer|SaveFile|Read Error|%v|%s", err, common.ToStringWithoutError(file))
			return err
		}
		_, err = openFile.Write(buf[:n])
	}
	logx.GetLogger("study").Infof("LambdaQueueServer|SaveFile|SaveFile Success|%s", targetPath)
	return nil
}

func Exec(ctx context.Context, name string, params ...string) error {
	cmd := exec.CommandContext(ctx, name, params...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logx.GetLogger("study").Errorf("Exec|StdoutPipe Error|%v", err)
		return err
	}

	if err := cmd.Start(); err != nil {
		logx.GetLogger("study").Errorf("Exec|Start Error|%v", err)
		return err
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	if err := cmd.Wait(); err != nil {
		logx.GetLogger("study").Errorf("Exec|Wait Error|%v", err)
		return err
	}
	return nil
}
