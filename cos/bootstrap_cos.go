package main

import (
	"context"
	"cos/api/config"
	http2 "cos/api/http"
	"cos/api/rpc"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	S3config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/redis/go-redis/v9"
)

func main() {

	//var configPath = flag.String("c", "E:\\code\\Go\\online_study_server\\cos\\api\\config\\config.json", "config.json file path")
	var configPath = flag.String("c", "/Users/zhijundu/code/GolandProjects/online_study_server/cos/api/config/config.json", "config.json file path")

	err := config.GetGloableConfig(*configPath)
	if err != nil {
		panic(err)
	}

	cfg, err := S3config.LoadDefaultConfig(
		context.TODO(),
		S3config.WithRegion(*config.GloableConfig.S3.Region),
		S3config.WithBaseEndpoint(*config.GloableConfig.S3.Endpoint),
		S3config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(*config.GloableConfig.S3.AccessKey, *config.GloableConfig.S3.SecretKey, "")))

	if err != nil {
		logx.GetLogger("study").Errorf("NewCosServer|S3config.LoadDefaultConfig err:%v", err)
		return
	}

	opt, err := redis.ParseURL(*config.GloableConfig.Redis)
	if err != nil {
		panic(err)
	}
	client := redis.NewClient(opt)
	hcli := &http.Client{
		Timeout: 30 * time.Second,
	}
	s3Client := s3.NewFromConfig(cfg, func(options *s3.Options) {
		options.HTTPClient = hcli
		options.UsePathStyle = true
		options.DisableLogOutputChecksumValidationSkipped = true
	})
	ctx, cancel := context.WithCancel(context.Background()) // 创建上下文
	defer cancel()
	// 启动http服务
	go http2.StartCosHttpServer(ctx, client, s3Client)
	// 启动rpc服务
	go rpc.StartCosRpcServer(ctx, client, s3Client)

	// 创建信号通道
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// 监听信号
	go func() {
		<-signalChan
		logx.GetLogger("study").Info("main|Received shutdown signal, shutting down...")
		cancel()
	}()

	// 主 Goroutine 阻塞，等待取消信号
	<-ctx.Done()
	logx.GetLogger("study").Info("main|Shutdown complete")
}
