package main

import (
	"context"
	"cos/api/config"
	http2 "cos/api/http"
	"cos/api/rpc"
	"flag"
	S3config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
	"net/http"
	"strconv"
	"time"
)

func main() {

	var configPath = flag.String("c", "E:\\code\\Go\\online_study_server\\cos\\api\\config\\config.json", "config.json file path")
	//var configPath = flag.String("c", "/Users/zhijundu/code/GolandProjects/online_study_server/cos/api/config/config.json", "config.json file path")
	err := config.RefreshEtcdConfig(*configPath)
	if err != nil {
		logx.GetLogger("study").Errorf("apiService|RefreshEtcdConfig|err:%v", err)
		return
	}

	err = config.LoadConfigFromEtcd()
	if err != nil {
		logx.GetLogger("study").Errorf("apiService|LoadConfigFromEtcd|err:%v", err)
		return
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

	client := redis.NewClient(&redis.Options{
		Addr:     *config.GloableConfig.Redis.Host + ":" + strconv.Itoa(*config.GloableConfig.Redis.Port),
		Username: *config.GloableConfig.Redis.Username,
		Password: *config.GloableConfig.Redis.Password,
		DB:       *config.GloableConfig.Redis.DB,
	})
	hcli := &http.Client{
		Timeout: 30 * time.Second,
	}
	s3Client := s3.NewFromConfig(cfg, func(options *s3.Options) {
		options.HTTPClient = hcli
		options.UsePathStyle = true
	})
	ctx, cancel := context.WithCancel(context.Background()) // 创建上下文
	defer cancel()                                          // 确保在退出时取消所有子任务
	var g errgroup.Group
	g.Go(func() error {
		err := http2.StartCosHttpServer(ctx, client, s3Client)
		if err != nil {
			logx.GetLogger("study").Errorf("main|StartApiServer|err:%v", err)
			cancel()
			return err
		}
		return nil
	})
	g.Go(func() error {
		err := rpc.StartCosRpcServer(ctx)
		if err != nil {
			logx.GetLogger("study").Errorf("main|StartRpcServer|err:%v", err)
			cancel()
			return err
		}
		return nil
	})
	if err := g.Wait(); err != nil {
		logx.GetLogger("study").Errorf("main|err:%v", err)
	}
}
