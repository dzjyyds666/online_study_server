package config

import (
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	S3 S3Config
}

type S3Config struct {
	Buckets   []string `json:"buckets"`
	Region    string   `json:"region"`
	AccessKey string   `json:"accessKey"`
	SecretKey string   `json:"secretKey"`
	Endpoint  string   `json:"endpoint"`
}
