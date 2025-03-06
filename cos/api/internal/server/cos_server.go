package server

import (
	"context"
	"cos/api/config"
	"cos/api/internal/core"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
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

type CosServer struct {
	s3Client *s3.Client
	hcli     *http.Client
	redis    *redis.Client
	bucket   string
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
		bucket:   aws.ToString(config.GloableConfig.S3.Bucket[0]),
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

	decoder := json.NewDecoder(ctx.Request().Body)
	var cosFile core.CosFile
	err := decoder.Decode(&cosFile)
	if err != nil {
		logx.GetLogger("OS_Server").Infof("HandlerApplyUpload|Params bind Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	// uuid生成fid
	fid := core.GenerateFid()
	cosFile.WithFid(fid)

	err = cosFile.CraetePrepareIndex(ctx, cs.redis)
	if err != nil && !errors.Is(err, core.ErrPrepareIndexExits) {
		logx.GetLogger("OS_Server").Errorf("HandlerApplyUpload|CraetePrepareIndex err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, "System Error")
	} else if errors.Is(err, core.ErrPrepareIndexExits) {
		// 获取redis中的数据
		if err != nil {
			logx.GetLogger("OS_Server").Errorf("HandlerApplyUpload|Get err:%v", err)
			return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
				"msg": err.Error(),
			})
		}
	}
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, cosFile)
}

func (cs *CosServer) HandlerSingleUpload(ctx echo.Context) error {

	fid := ctx.Param("fid")
	dirId := ctx.QueryParam("dirid")
	file, err := ctx.FormFile("file")
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerSingleUpload|FormFile err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	filename := file.Filename
	fileSize := file.Size

	open, err := file.Open()
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerSingleUpload|Open err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Open File Error",
		})
	}
	defer open.Close()

	// 计算文件的md5值
	md5, err := core.CalculateMD5(open)
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerSingleUpload|CalculateMD5 err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "CalculateMD5 Error",
		})
	}

	open.Seek(0, 0)

	var uploadFile core.CosFile

	uploadFile.WithFid(fid).
		WithDirectoryId(dirId).
		WithFileName(filename).
		WithFileSize(fileSize).
		WithReader(open).
		WithFileMD5(md5)

	// 校验文件是否与初始化上传文件一致
	index, err := core.QueryPrepareIndex(ctx, cs.redis, aws.ToString(uploadFile.Fid))
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerSingleUpload|QueryPrepareIndex err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Query PrepareIndex Error",
		})
	}

	if !index.IsMatch(&uploadFile) {
		logx.GetLogger("OS_Server").Errorf("HandlerSingleUpload|QueryPrepareIndex|Not Match")
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpBadRequest, echo.Map{
			"msg": "File Not Match",
		})
	}

	err = uploadFile.PutObject(ctx, cs.s3Client, aws.String(cs.bucket))
	if nil != err {
		logx.GetLogger("OS_Server").Errorf("HandlerSingleUpload|PutObject err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpBadRequest, echo.Map{
			"msg": "PutObject Error",
		})
	}

	err = index.CraeteIndex(ctx, cs.redis)
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerSingleUpload|CraeteIndex err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "CraeteIndex Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, uploadFile)
}

func (cs *CosServer) HandlerGetFile(ctx echo.Context) error {
	// 获取文件,直接拼接url访问minio

	fid := ctx.Param("fid")
	index, err := core.QueryIndex(ctx, cs.redis, fid)
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerGetFile|QueryIndex err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpBadRequest, echo.Map{
			"msg": "Index File Not Exist",
		})
	}

	path := index.GetFilePath()

	url := aws.ToString(config.GloableConfig.S3.Endpoint) + "/" + cs.bucket + path

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerGetFile|NewRequest err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "NewRequest err",
		})
	}

	resp, err := cs.hcli.Do(request)
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerGetFile|Do err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Get File Error",
		})
	}
	defer resp.Body.Close()

	return ctx.Stream(http.StatusOK, aws.ToString(index.FileType), resp.Body)
}

func (cs *CosServer) HandlerInitMultipartUpload(ctx echo.Context) error {
	var initupload core.InitMultipartUpload
	err := ctx.Bind(&initupload)
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerInitMultipartUpload|Bind err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	initupload.WithBucket(cs.bucket)

	// 初始化上传文件
	err = initupload.InitUpload(ctx, cs.s3Client)

	return nil
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
