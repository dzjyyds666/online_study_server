package main

import (
	"class/api/config"
	"class/api/http"
	"class/api/rpc"
	"context"
	"flag"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/sync/errgroup"
)

func main() {
	var configPath = flag.String("c", "/Users/aaron/GolandProjects/online_study_server/class/api/config/config.json", "config.json file path")
	//var configPath = flag.String("c", "E:\\code\\Go\\online_study_server\\class\\api\\config\\config.json", "config.json file path")
	//err := config.RefreshEtcdConfig(*configPath)
	//if err != nil {
	//	panic(err)
	//}
	//err = config.LoadConfigFromEtcd()
	//if err != nil {
	//	panic(err)
	//}
	err := config.GetGloableConfig(*configPath)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	url, err := redis.ParseURL(config.GloableConfig.Redis)
	if err != nil {
		panic(err)
	}

	dsClient := redis.NewClient(url)
	//连接mongodb
	mgDb, err := mongo.Connect(ctx, options.Client().ApplyURI(config.GloableConfig.Mongo))
	if err != nil {
		panic(err)
	}

	var g errgroup.Group
	g.Go(func() error {
		err := http.StartClassHttpServer(ctx, dsClient, mgDb)
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
