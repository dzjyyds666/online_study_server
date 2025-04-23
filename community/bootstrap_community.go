package main

import (
	"community/api/config"
	"community/api/core"
	"community/api/http"
	"community/api/rpc"
	"context"
	"flag"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var configPath = flag.String("c", "./api/config/config.json", "config.json file path")
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	opt, err := redis.ParseURL(config.GloableConfig.Redis)
	if err != nil {
		panic(err)
	}

	// 连接redis
	dsClient := redis.NewClient(opt)

	// mongoDb
	mgDb, err := mongo.Connect(ctx, options.Client().ApplyURI(config.GloableConfig.Mongo))
	if err != nil {
		panic(err)
	}

	plateServer := core.NewPlateServer(ctx, dsClient, mgDb)

	go http.StartCommunityHttpServer(ctx, plateServer)

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
