package svc

import (
	"context"
	S3config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/dzjyyds666/online_study_server/object_storage/api/internal/config"
	"github.com/dzjyyds666/online_study_server/object_storage/api/internal/middleware"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/zeromicro/go-zero/rest"
	"net/http"
	"time"
)

type ServiceContext struct {
	Config         config.Config
	AuthMiddleware rest.Middleware
	S3             *s3.Client
	Hcli           *http.Client
}

func NewServiceContext(c config.Config) *ServiceContext {

	// 初始化s3客户端
	cfg, err := S3config.LoadDefaultConfig(
		context.TODO(),
		S3config.WithRegion(c.S3.Region),
		S3config.WithBaseEndpoint(c.S3.Endpoint),
		S3config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(c.S3.AccessKey, c.S3.SecretKey, "")))
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("S3config.LoadDefaultConfig err:%v", err)
		panic(err)
	}

	s3Client := s3.NewFromConfig(cfg, func(options *s3.Options) {
		options.HTTPClient = &http.Client{Timeout: 30 * time.Second}
		options.UsePathStyle = true
	})

	return &ServiceContext{
		Config:         c,
		AuthMiddleware: middleware.NewAuthMiddleware().Handle,
		Hcli: &http.Client{
			Timeout: 30 * time.Second,
		},
		S3: s3Client,
	}
}
