package config

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/zeromicro/go-zero/rest"
	"io"
	"os"
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
	Policy    string   `json:"policy"`
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

			_, err = s3Client.PutBucketPolicy(context.TODO(), &s3.PutBucketPolicyInput{
				Bucket: aws.String(bucket),
				Policy: aws.String(sc.Policy),
			})
			if nil != err {
				logx.GetLogger("OS_Server").Errorf("s3Client|%s|PutBucketPolicy err:%v", bucket, err)
				panic(err)
			}
		}
	}
}

func (sc *S3Config) WithPolicy(policyFile string) (*S3Config, error) {
	open, err := os.Open(policyFile)
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("S3Config.WithPolicy|open err:%v", err)
		return nil, err
	}
	defer open.Close()

	policy, err := io.ReadAll(open)
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("S3Config.WithPolicy|readAll err:%v", err)
		return nil, err
	}

	sc.Policy = string(policy)
	return sc, nil
}
