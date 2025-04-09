package core

import (
	"context"
	"cos/api/config"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/dzjyyds666/opensource/common"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/redis/go-redis/v9"
	"io"
	"net/http"
	"os"
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

func (lqs *LambdaQueueServer) WatchQueue(ctx context.Context, queueName string) error {
	// 开始监听队列
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
	objectKey := file.MergeFilePath()
	// 从客户端中拉取下来文件
	r, err := lqs.cosServ.GetFile(ctx, lqs.bucket, objectKey)
	if err != nil {
		logx.GetLogger("study").Errorf("LambdaQueueServer|FormatVideo|GetFile Error|%v|%s", err, common.ToStringWithoutError(file))
		return err
	}
	defer r.Close()
	// 把文件写入到临时路径
	err = lqs.SaveFile(ctx, file, r)
	if err != nil {
		logx.GetLogger("study").Errorf("LambdaQueueServer|FormatVideo|SaveFile Error|%v|%s", err, common.ToStringWithoutError(file))
		return err
	}

	// todo 执行文件的处理

	return nil
}

func (lqs *LambdaQueueServer) SaveFile(ctx context.Context, file *CosFile, reader io.Reader) error {
	targetPath := path.Join(lqs.tmpDir, aws.ToString(file.FileName))
	openFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0642)
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
