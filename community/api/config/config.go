package config

import (
	"context"
	"encoding/json"
	"github.com/dzjyyds666/opensource/logx"
	clientv3 "go.etcd.io/etcd/client/v3"
	"io"
	"os"
	"time"
)

// 全局配置
var GloableConfig Config

type Config struct {
	Port    *int    `json:"port"`     // http服务端口
	RpcPort *int    `json:"rpc_port"` // rpc服务端口
	Host    *string `json:"host"`     // 服务器地址
	Name    *string `json:"name"`     // 服务器名称

	Mysql Mysql `json:"mysql"`
	Redis Redis `json:"redis"`
	Jwt   Jwt   `json:"jwt"`
	Mongo Mongo `json:"mongo"`
}

type Mysql struct {
	Host     *string `json:"host"`
	Port     *int    `json:"port"`
	Username *string `json:"username"`
	Password *string `json:"password"`
	Database *string `json:"database"`
}

type Redis struct {
	Host     *string `json:"host"`
	Port     *int    `json:"port"`
	Username *string `json:"username"`
	Password *string `json:"password"`
	DB       *int    `json:"db"`
}

type Jwt struct {
	Secretkey *string `json:"secret_key"`
	Expire    *int    `json:"expire"`
}

type Mongo struct {
	Host     *string `json:"host"`
	Port     *int    `json:"port"`
	Username *string `json:"username"`
	Password *string `json:"password"`
	DB       *string `json:"database"`
}

func LoadConfigFromEtcd() error {
	// 从etcd中加载配置
	client, err := clientv3.New(clientv3.Config{Endpoints: []string{"127.0.0.1:2379"}, DialTimeout: 5 * time.Second})
	if err != nil {
		logx.GetLogger("study").Errorf("LoadConfigFromEtcd|clientv3.New err:%v", err)
		return err
	}
	defer client.Close()

	// 使用ctx控制超时
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	cfg, err := client.Get(ctx, "user.proto:config.json")
	cancel()
	if nil != err {
		logx.GetLogger("study").Errorf("LoadConfigFromEtcd|client.Get err:%v", err)
		return err
	}
	err = json.Unmarshal(cfg.Kvs[0].Value, &GloableConfig)
	if nil != err {
		logx.GetLogger("study").Errorf("LoadConfigFromEtcd|json.Unmarshal err:%v", err)
		return err
	}

	//logx.GetLogger("study").Infof("LoadConfigFromEtcd|SUCC|GloableConfig|%v", console.ToStringWithoutError(GloableConfig))
	return nil
}

func RefreshEtcdConfig(path string) error {
	// 从文件中读取到配置，写入etcd
	open, err := os.Open(path)
	if err != nil {
		return err
	}

	defer open.Close()
	all, err := io.ReadAll(open)
	if err != nil {
		return err
	}

	jsonConfig := string(all)

	// 从etcd中加载配置
	client, err := clientv3.New(clientv3.Config{Endpoints: []string{"127.0.0.1:2379"}, DialTimeout: 5 * time.Second})
	if err != nil {
		logx.GetLogger("study").Errorf("RefreshEtcdConfig|clientv3.New err:%v", err)
		return err
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	_, err = client.Put(ctx, "user.proto:config.json", jsonConfig)
	cancel()
	if err != nil {
		logx.GetLogger("study").Errorf("RefreshEtcdConfig|client.Put err:%v", err)
		return err
	}

	//logx.GetLogger("study").Infof("RefreshEtcdConfig|SUCC|GloableConfig|%v", console.ToStringWithoutError(GloableConfig))
	return nil
}
