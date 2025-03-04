package server

import (
	"context"
	"cos/api/config"
	"cos/api/internal/core"
	"errors"
	"fmt"
	S3config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/dzjyyds666/opensource/httpx"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/labstack/echo"
	"github.com/redis/go-redis/v9"
	"net/http"
	"strconv"
	"time"
)

var defaultBucket = config.GloableConfig.S3.Bucket[0]

type CosServer struct {
	s3Client *s3.Client
	hcli     *http.Client
	redis    *redis.Client
}

func NewCosServer() (*CosServer, error) {
	// 初始化s3客户端
	cfg, err := S3config.LoadDefaultConfig(
		context.TODO(),
		S3config.WithRegion(*config.GloableConfig.S3.Region),
		S3config.WithBaseEndpoint(*config.GloableConfig.S3.Endpoint),
		S3config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(*config.GloableConfig.S3.AccessKey, *config.GloableConfig.S3.SecretKey, "")))
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("NewCosServer|S3config.LoadDefaultConfig err:%v", err)
		return nil, err
	}

	client := redis.NewClient(&redis.Options{
		Addr:     *config.GloableConfig.Redis.Host + ":" + strconv.Itoa(*config.GloableConfig.Redis.Port),
		Username: *config.GloableConfig.Redis.Username,
		Password: *config.GloableConfig.Redis.Password,
		DB:       *config.GloableConfig.Redis.DB,
	})

	hcli := &http.Client{Timeout: 30 * time.Second}

	s3Client := s3.NewFromConfig(cfg, func(options *s3.Options) {
		options.HTTPClient = &http.Client{Timeout: 30 * time.Second}
		options.UsePathStyle = true
	})

	cosServer := &CosServer{
		s3Client: s3Client,
		hcli:     hcli,
		redis:    client,
	}

	err = cosServer.checkAndCreateBucket()
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("NewCosServer|CheckAndCreateBucket|err:%v", err)
		return nil, err
	}

	return cosServer, nil
}

// 需要传入文件的名称，md5，长度，文件夹id
func (cs *CosServer) HandlerApplyUpload(ctx echo.Context) error {
	var cosFile *core.CosFile
	err := ctx.Bind(cosFile)
	if err != nil {
		logx.GetLogger("OS_Server").Infof("HandlerApplyUpload|Params bind Error|%v", err)
		return ctx.JSON(http.StatusBadRequest, httpx.HttpResponse{
			StatusCode: httpx.HttpStatusCode.HttpBadRequest,
			Msg:        "Params Invalid",
		})
	}

	// uuid生成fid
	fid := core.GenerateFid()
	cosFile.WithFid(fid)

	err = cosFile.CraetePrepareIndex(ctx, cs.redis)
	if err != nil && !errors.Is(err, core.ErrPrepareIndexExits) {
		logx.GetLogger("OS_Server").Errorf("HandlerApplyUpload|CraetePrepareIndex err:%v", err)
		return ctx.JSON(http.StatusBadRequest, httpx.HttpResponse{
			StatusCode: httpx.HttpStatusCode.HttpInternalError,
			Msg:        "system error",
		})
	} else if errors.Is(err, core.ErrPrepareIndexExits) {
		// 获取redis中的数据
		if err != nil {
			logx.GetLogger("OS_Server").Errorf("HandlerApplyUpload|Get err:%v", err)
			return ctx.JSON(http.StatusBadRequest,
				httpx.HttpResponse{StatusCode: httpx.HttpStatusCode.HttpBadRequest, Msg: err.Error()})
		}
	}
	return ctx.JSON(http.StatusOK, httpx.HttpResponse{StatusCode: httpx.HttpStatusCode.HttpOK, Msg: "ok", Data: fid})
}

func (cs *CosServer) HandlerSingleUpload(ctx echo.Context) error {

	fid := ctx.Param("fid")
	dirId := ctx.Param("dirid")
	file, err := ctx.FormFile("file")
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerSingleUpload|FormFile err:%v", err)
		return ctx.JSON(http.StatusBadRequest,
			httpx.HttpResponse{StatusCode: httpx.HttpStatusCode.HttpBadRequest, Msg: err.Error()})
	}

	filename := file.Filename
	fileSize := file.Size

	open, err := file.Open()
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerSingleUpload|Open err:%v", err)
		return ctx.JSON(http.StatusBadRequest,
			httpx.HttpResponse{StatusCode: httpx.HttpStatusCode.HttpBadRequest, Msg: "System Error"})
	}
	defer open.Close()

	open.Seek(0, 0)

	//uploadFile := &core.CosFile{
	//	DirectoryId: &dirId,
	//	FileMD5:     &md5,
	//	FileName:    &filename,
	//	FileSize:    &fileSize,
	//	Fid:         &fid,
	//}

	var uploadFile *core.CosFile

	uploadFile.WithFid(fid).
		WithDirectoryId(dirId).
		WithFileName(filename).
		WithFileSize(fileSize).
		WithReader(open)

	md5, err := uploadFile.CalculateMD5()
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerSingleUpload|CalculateMD5 err:%v", err)
		return ctx.JSON(http.StatusBadRequest,
			httpx.HttpResponse{StatusCode: httpx.HttpStatusCode.HttpBadRequest, Msg: "System Error"})
	}

	// 校验文件是否与初始化上传文件一致
	index, err := core.QueryPrepareIndex(ctx, cs.redis, uploadFile)
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerSingleUpload|QueryPrepareIndex err:%v", err)
		return ctx.JSON(http.StatusBadRequest,
			httpx.HttpResponse{StatusCode: httpx.HttpStatusCode.HttpBadRequest, Msg: "Get PerpareIndex Error"})
	}

	if !index.IsMatch(uploadFile) {
		logx.GetLogger("OS_Server").Errorf("HandlerSingleUpload|QueryPrepareIndex|Not Match")
		return ctx.JSON(http.StatusBadRequest,
			httpx.HttpResponse{StatusCode: httpx.HttpStatusCode.HttpBadRequest, Msg: "File is Invalid"})
	}

	uploadFile.PutObject(ctx, cs.s3Client, defaultBucket)

	if nil != err {
		logx.GetLogger("OS_Server").Errorf("HandlerSingleUpload|PutObject err:%v", err)
		return ctx.JSON(http.StatusBadRequest,
			httpx.HttpResponse{StatusCode: httpx.HttpStatusCode.HttpBadRequest, Msg: "Upload Failed"})
	}

	err = index.CraeteIndex(ctx, cs.redis)
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerSingleUpload|CraeteIndex err:%v", err)
		return ctx.JSON(http.StatusBadRequest,
			httpx.HttpResponse{StatusCode: httpx.HttpStatusCode.HttpBadRequest, Msg: "System Error"})
	}

	return ctx.JSON(http.StatusOK, httpx.HttpResponse{StatusCode: httpx.HttpStatusCode.HttpOK, Msg: "Upload Success", Data: fid})
}

func (cs *CosServer) checkAndCreateBucket() error {
	for _, bucket := range config.GloableConfig.S3.Bucket {
		_, err := cs.s3Client.HeadBucket(context.TODO(), &s3.HeadBucketInput{
			Bucket: bucket,
		})
		if err != nil {
			logx.GetLogger("OS_Server").Errorf("CheckAndCreateBucket|HeadBucket err:%v", err)

			// 创建bucket
			_, err = cs.s3Client.CreateBucket(context.TODO(), &s3.CreateBucketInput{
				Bucket: bucket,
			})
			if err != nil {
				logx.GetLogger("OS_Server").Errorf("CheckAndCreateBucket|CreateBucket err:%v", err)
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
				logx.GetLogger("OS_Server").Errorf("CheckAndCreateBucket|PutBucketPolicy err:%v", err)
				return err
			}
		}
	}
	return nil
}
