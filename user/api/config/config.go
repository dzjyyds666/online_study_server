package config

import (
	"context"
	"encoding/json"
	"github.com/dzjyyds666/opensource/logx"
	clientv3 "go.etcd.io/etcd/client/v3"
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

func LoadConfigFromEtcd() error {
	// 从etcd中加载配置
	client, err := clientv3.New(clientv3.Config{Endpoints: []string{"127.0.0.1:2379"}, DialTimeout: 5 * time.Second})
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("LoadConfigFromEtcd|clientv3.New err:%v", err)
		return err
	}
	defer client.Close()

	// 使用ctx控制超时
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	cfg, err := client.Get(ctx, "user:config")
	cancel()
	if nil != err {
		logx.GetLogger("OS_Server").Errorf("LoadConfigFromEtcd|client.Get err:%v", err)
		return err
	}
	err = json.Unmarshal(cfg.Kvs[0].Value, &GloableConfig)
	if nil != err {
		logx.GetLogger("OS_Server").Errorf("LoadConfigFromEtcd|json.Unmarshal err:%v", err)
		return err
	}

	//logx.GetLogger("OS_Server").Infof("LoadConfigFromEtcd|SUCC|GloableConfig|%v", common.ToStringWithoutError(GloableConfig))
	return nil
}

func RefreshEtcdConfig(jsonConfig string) error {
	// 从etcd中加载配置
	client, err := clientv3.New(clientv3.Config{Endpoints: []string{"127.0.0.1:2379"}, DialTimeout: 5 * time.Second})
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("RefreshEtcdConfig|clientv3.New err:%v", err)
		return err
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	_, err = client.Put(ctx, "user:config", jsonConfig)
	cancel()
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("RefreshEtcdConfig|client.Put err:%v", err)
		return err
	}

	//logx.GetLogger("OS_Server").Infof("RefreshEtcdConfig|SUCC|GloableConfig|%v", common.ToStringWithoutError(GloableConfig))
	return nil
}
