package server

import (
	"context"
	"cos/api/config"
	"fmt"
	S3config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/dzjyyds666/opensource/logx"
	"net/http"
	"time"
)

type CosServer struct {
	s3Client *s3.Client
	hcli     *http.Client
}

func NewCosServer() (*CosServer, error) {
	// 初始化s3客户端
	cfg, err := S3config.LoadDefaultConfig(
		context.TODO(),
		S3config.WithRegion(*config.GloableConfig.S3.Region),
		S3config.WithBaseEndpoint(*config.GloableConfig.S3.Endpoint),
		S3config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(*config.GloableConfig.S3.AccessKey, *config.GloableConfig.S3.SecretKey, "")))
	if err != nil {
		logx.GetLogger("COS_Server").Errorf("NewCosServer|S3config.LoadDefaultConfig err:%v", err)
		return nil, err
	}

	hcli := &http.Client{Timeout: 30 * time.Second}

	s3Client := s3.NewFromConfig(cfg, func(options *s3.Options) {
		options.HTTPClient = &http.Client{Timeout: 30 * time.Second}
		options.UsePathStyle = true
	})

	cosServer := &CosServer{
		s3Client: s3Client,
		hcli:     hcli,
	}

	err = cosServer.CheckAndCreateBucket()
	if err != nil {
		logx.GetLogger("COS_Server").Errorf("NewCosServer|CheckAndCreateBucket|err:%v", err)
		return nil, err
	}

	return cosServer, nil
}

func (cs *CosServer) CheckAndCreateBucket() error {
	for _, bucket := range config.GloableConfig.S3.Bucket {
		_, err := cs.s3Client.HeadBucket(context.TODO(), &s3.HeadBucketInput{
			Bucket: bucket,
		})
		if err != nil {
			logx.GetLogger("COS_Server").Errorf("CheckAndCreateBucket|HeadBucket err:%v", err)

			// 创建bucket
			_, err = cs.s3Client.CreateBucket(context.TODO(), &s3.CreateBucketInput{
				Bucket: bucket,
			})
			if err != nil {
				logx.GetLogger("COS_Server").Errorf("CheckAndCreateBucket|CreateBucket err:%v", err)
				return err
			}

			// 设置桶的策略
			policy := `
				{
     "Version": "2012-10-17",
     "Statement": [
       {
         "Effect": "Allow",
         "Principal": "*",
         "Action": "s3:GetObject",
         "Resource": "arn:aws:s3:::%s/*"
       }
     ]
   }
`
			policy = fmt.Sprintf(policy, *bucket)
			_, err = cs.s3Client.PutBucketPolicy(context.TODO(), &s3.PutBucketPolicyInput{
				Bucket: bucket,
				Policy: &policy,
			})
			if err != nil {
				logx.GetLogger("COS_Server").Errorf("CheckAndCreateBucket|PutBucketPolicy err:%v", err)
				return err
			}
		}
	}
	return nil
}
