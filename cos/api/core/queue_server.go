package core

import (
	"bufio"
	"context"
	"cos/api/config"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/dzjyyds666/opensource/common"
	"github.com/dzjyyds666/opensource/logx"
	"io"
	"net/http"
	"os"
	"os/exec"
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
