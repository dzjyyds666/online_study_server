package main

import (
	"community/api/config"
	"community/api/http"
	"community/api/rpc"
	"context"
	"flag"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/redis/go-redis/v9"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

func main() {
	var configPath = flag.String("c", "/Users/zhijundu/code/GolandProjects/online_study_server/class/api/config/config.json", "config.json file path")
	//var configPath = flag.String("c", "E:\\code\\Go\\online_study_server\\class\\api\\config\\config.json", "config.json file path")
	err := config.RefreshEtcdConfig(*configPath)
	if err != nil {
		logx.GetLogger("study").Errorf("main|RefreshEtcdConfig|err:%v", err)
		return
	}

	err = config.LoadConfigFromEtcd()
	if err != nil {
		panic(err)
	}

	// 连接redis
	dsClient := redis.NewClient(&redis.Options{
		Addr:     *config.GloableConfig.Redis.Host + ":" + strconv.Itoa(*config.GloableConfig.Redis.Port),
		Username: *config.GloableConfig.Redis.Username,
		Password: *config.GloableConfig.Redis.Password,
		DB:       *config.GloableConfig.Redis.DB,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go http.StartCommunityHttpServer(ctx, dsClient)

	go rpc.StartCommunityRpcService(ctx)

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
