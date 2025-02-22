package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	Redis RedisConfig
	Mysql MySqlConfig
}

type RedisConfig struct {
	Host     string `json:"host"`
	Password string `json:"password"`
	DB       int    `json:"db"`
	Port     int    `json:"port"`
}

type MySqlConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	UserName string `json:"username"`
	DB       string `json:"db"`
}
