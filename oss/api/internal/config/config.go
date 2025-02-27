package config

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/dzjyyds666/opensource/logx"
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
	Policy    Policy   `json:"policy"`
}

type Policy struct {
	Version   string      `json:"Version"`
	Statement []Statement `json:"Statement"`
}

type Statement struct {
	Effect    string `json:"Effect"`
	Principal string `json:"Principal"`
	Action    string `json:"Action"`
	Resource  string `json:"Resource"`
}

func (sc *S3Config) CheckAndCreateBucket(s3Client *s3.Client) {
	// 检查bucket是否存在，不存在的话创建该bucket
	for _, bucket := range sc.Buckets {
		_, err := s3Client.HeadBucket(context.TODO(), &s3.HeadBucketInput{Bucket: &bucket})
		if nil != err {
			logx.GetLogger("OS_Server").Warnf("s3Client|%s|HeadBucket err:%v", bucket, err)
			// 创建bucket
			_, err = s3Client.CreateBucket(context.TODO(), &s3.CreateBucketInput{
				Bucket: aws.String(bucket),
			})
			if nil != err {
				logx.GetLogger("OS_Server").Errorf("s3Client|%s|CreateBucket err:%v", bucket, err)
				panic(err)
			}

			raw, err := json.Marshal(sc.Policy)
			if nil != err {
				logx.GetLogger("OS_Server").Infof("CheckAndCreateBucket|JSON Marshal Error|%v", err)
				panic(err)
			}

			_, err = s3Client.PutBucketPolicy(context.TODO(), &s3.PutBucketPolicyInput{
				Bucket: aws.String(bucket),
				Policy: aws.String(fmt.Sprintf(string(raw), bucket)),
			})
			if nil != err {
				logx.GetLogger("OS_Server").Errorf("s3Client|%s|PutBucketPolicy err:%v", bucket, err)
				panic(err)
			}
		}
	}
}
