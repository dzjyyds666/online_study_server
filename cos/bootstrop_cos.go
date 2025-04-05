package main

import (
	"context"
	"cos/api/config"
	"cos/api/grpc"
	"cos/api/http"
	"flag"
	"github.com/dzjyyds666/opensource/logx"
	"golang.org/x/sync/errgroup"
)

func main() {

	var configPath = flag.String("c", "/Users/zhijundu/code/GolandProjects/online_study_server/cos/api/config/config.json", "config.json file path")
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

	ctx, cancel := context.WithCancel(context.Background()) // 创建上下文
	defer cancel()                                          // 确保在退出时取消所有子任务

	var g errgroup.Group

	g.Go(func() error {
		err := http.StartCosHttpServer(ctx)
		if err != nil {
			logx.GetLogger("study").Errorf("main|StartApiServer|err:%v", err)
			cancel()
			return err
		}
		return nil
	})

	g.Go(func() error {
		err := grpc.StartCosRpcServer(ctx)
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
