package main

import (
	"class/api/config"
	"class/api/http"
	"class/api/rpc"
	"context"
	"flag"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
	"strconv"
)

func main() {
	//var configPath = flag.String("c", "/Users/zhijundu/code/GolandProjects/online_study_server/class/api/config/config.json", "config.json file path")
	var configPath = flag.String("c", "E:\\code\\Go\\online_study_server\\class\\api\\config\\config.json", "config.json file path")
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

	var g errgroup.Group
	g.Go(func() error {
		err := http.StartClassHttpServer(ctx, dsClient)
		if nil != err {
			logx.GetLogger("study").Errorf("main|StartApiServer|err:%v", err)
			cancel()
			return err
		}
		return nil
	})
	g.Go(func() error {
		err := rpc.StratClassRpcService(ctx, dsClient)
		if nil != err {
			logx.GetLogger("study").Errorf("main|StartRpcServer|err:%v", err)
			rpc.StopClassRpcService()
			cancel()
			return err
		}
		return nil
	})
	if err := g.Wait(); err != nil {
		logx.GetLogger("study").Errorf("main|err:%v", err)
	}
}
