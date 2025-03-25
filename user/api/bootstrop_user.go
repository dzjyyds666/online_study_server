package main

import (
	"context"
	"flag"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strconv"
	"user/api/config"
	"user/api/http"
	"user/api/rpc"
)

func main() {
	var configPath = flag.String("c", "/Users/zhijundu/GolandProjects/online_study_server/user/api/config/config.json", "config.json file path")
	err := config.RefreshEtcdConfig(*configPath)
	if err != nil {
		logx.GetLogger("study").Errorf("main|RefreshEtcdConfig|err:%v", err)
		return
	}

	err = config.LoadConfigFromEtcd()
	if err != nil {
		panic(err)
	}

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

	defer cancel()

	var g errgroup.Group

	g.Go(func() error {
		err = http.StartUserHttpServer(ctx, client, mysql)
		if err != nil {
			logx.GetLogger("study").Errorf("main|StartUserHttpServer|err:%v", err)
			cancel()
			return err
		}
		return nil
	})

	g.Go(func() error {
		err = rpc.StratUserRpcServer(ctx, client, mysql)
		if err != nil {
			logx.GetLogger("study").Errorf("main|StratUserRpcServer|err:%v", err)
			cancel()
			return err
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		logx.GetLogger("study").Errorf("main|err:%v", err)
		return
	}
}
