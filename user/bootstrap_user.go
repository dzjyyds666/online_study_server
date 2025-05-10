package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"user/api/config"
	"user/api/core"
	"user/api/http"
	"user/api/rpc"

	"github.com/dzjyyds666/opensource/common"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	var configPath = flag.String("c", "/Users/zhijundu/code/GolandProjects/online_study_server/user/api/config/config.json", "config.json file path")
	//var configPath = flag.String("c", "E:\\code\\Go\\online_study_server\\user\\api\\config\\config.json", "config.json file path")

	err := config.GetGloableConfig(*configPath)
	if err != nil {
		panic(err)
	}

	logx.GetLogger("study").Infof("main|LoadConfigFromEtcd|config:%s", common.ToStringWithoutError(config.GloableConfig))
	client := redis.NewClient(&redis.Options{
		Addr:     *config.GloableConfig.Redis.Host + ":" + strconv.Itoa(*config.GloableConfig.Redis.Port),
		Username: *config.GloableConfig.Redis.Username,
		Password: *config.GloableConfig.Redis.Password,
		DB:       *config.GloableConfig.Redis.DB,
	})
	// gorm连接mysql
	dsn := *config.GloableConfig.Mysql.Username + ":" + *config.GloableConfig.Mysql.Password + "@tcp(" + *config.GloableConfig.Mysql.Host + ":" + strconv.Itoa(*config.GloableConfig.Mysql.Port) + ")/" + *config.GloableConfig.Mysql.Database + "?charset=utf8mb4&parseTime=True&loc=Local"
	mysql, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if nil != err {
		logx.GetLogger("study").Errorf("NewUserServer|gorm.Open err:%v", err)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	server, err := core.NewUserServer(ctx, client, mysql)
	if err != nil {
		logx.GetLogger("study").Errorf("main|NewUserServer|err:%v", err)
		return
	}
	defer cancel()

	go http.StartUserHttpServer(ctx, server)

	go rpc.StratUserRpcServer(ctx, server)

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
