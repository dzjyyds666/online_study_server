package server

import (
	"context"
	"cos/api/config"
	core2 "cos/api/http/internal/core"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	S3config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/dzjyyds666/opensource/common"
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
		logx.GetLogger("study").Errorf("NewCosServer|S3config.LoadDefaultConfig err:%v", err)
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
		logx.GetLogger("study").Errorf("NewCosServer|CheckAndCreateBucket|err:%v", err)
		return nil, err
	}

	return cosServer, nil
}

// 需要传入文件的名称，md5，长度，文件夹id
func (cs *CosServer) HandlerApplyUpload(ctx echo.Context) error {

	decoder := json.NewDecoder(ctx.Request().Body)
	var cosFile core2.CosFile
	err := decoder.Decode(&cosFile)
	if err != nil {
		logx.GetLogger("study").Infof("HandlerApplyUpload|Params bind Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	// uuid生成fid
	fid := core2.GenerateFid()
	cosFile.WithFid("fi_" + fid)

	err = cosFile.CraetePrepareIndex(ctx, cs.redis)
	if err != nil && !errors.Is(err, core2.ErrPrepareIndexExits) {
		logx.GetLogger("study").Errorf("HandlerApplyUpload|CraetePrepareIndex err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, "System Error")
	} else if errors.Is(err, core2.ErrPrepareIndexExits) {
		// 获取redis中的数据
		if err != nil {
			logx.GetLogger("study").Errorf("HandlerApplyUpload|Get err:%v", err)
			return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
				"msg": err.Error(),
			})
		}
	}
	logx.GetLogger("study").Infof("HandlerApplyUpload|CraetePrepareIndex|Succ|%s", common.ToStringWithoutError(cosFile))
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, cosFile)
}

func (cs *CosServer) HandlerSingleUpload(ctx echo.Context) error {

	fid := ctx.Param("fid")
	dirId := ctx.QueryParam("dirid")
	file, err := ctx.FormFile("file")
	if err != nil {
		logx.GetLogger("study").Errorf("HandlerSingleUpload|FormFile err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	logx.GetLogger("study").Infof("HandlerSingleUpload|UploadSingleFile|%s", fid)

	//查询redis中有没有预备的信息
	cosFile, err := core2.QueryPrepareIndex(ctx, cs.redis, fid)
	if err != nil {
		logx.GetLogger("study").Errorf("HandlerSingleUpload|QueryPrepareIndex err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "QueryPrepareIndex Error",
		})
	}

	if cosFile == nil {
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpBadRequest, echo.Map{
			"msg": "Prepare Index Not Exist",
		})
	}

	filename := file.Filename
	fileSize := file.Size

	open, err := file.Open()
	if err != nil {
		logx.GetLogger("study").Errorf("HandlerSingleUpload|Open err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Open File Error",
		})
	}
	defer open.Close()

	// 计算文件的md5值
	md5, err := core2.CalculateMD5(open)
	if err != nil {
		logx.GetLogger("study").Errorf("HandlerSingleUpload|CalculateMD5 err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "CalculateMD5 Error",
		})
	}
	open.Seek(0, 0)

	// 计算文件的type
	fileType, err := core2.GetFileType(open)
	if err != nil {
		logx.GetLogger("study").Errorf("HandlerSingleUpload|GetFileType err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "GetFileType Error",
		})
	}
	open.Seek(0, 0)

	var uploadFile core2.CosFile

	uploadFile.WithFid(fid).
		WithDirectoryId(dirId).
		WithFileName(filename).
		WithFileSize(fileSize).
		WithReader(open).
		WithFileMD5(md5).
		WithFileType(fileType)

	logx.GetLogger("study").Infof("HandlerSingleUpload|UploadSingleFile|Info|%s", common.ToStringWithoutError(uploadFile))

	err = uploadFile.UploadSingleFile(ctx, cs.s3Client, aws.String(cs.bucket), cs.redis)
	if nil != err {
		logx.GetLogger("study").Errorf("HandlerSingleUpload|PutObject err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpBadRequest, echo.Map{
			"msg": "PutObject Error",
		})
	}

	logx.GetLogger("study").Infof("HandlerSingleUpload|UploadSingleFile|Succ|%s", fid)

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, uploadFile)
}

func (cs *CosServer) HandlerGetFile(ctx echo.Context) error {
	// 获取文件,直接拼接url访问minio

	fid := ctx.Param("fid")
	index, err := core2.QueryIndex(ctx, cs.redis, fid)
	if err != nil {
		logx.GetLogger("study").Errorf("HandlerGetFile|QueryIndex err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpBadRequest, echo.Map{
			"msg": "Index File Not Exist",
		})
	}

	path := index.GetFilePath()

	url := aws.ToString(config.GloableConfig.S3.Endpoint) + "/" + cs.bucket + path

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logx.GetLogger("study").Errorf("HandlerGetFile|NewRequest err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "NewRequest err",
		})
	}

	resp, err := cs.hcli.Do(request)
	if err != nil {
		logx.GetLogger("study").Errorf("HandlerGetFile|Do err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Get File Error",
		})
	}
	defer resp.Body.Close()

	return ctx.Stream(http.StatusOK, aws.ToString(index.FileType), resp.Body)
}

func (cs *CosServer) HandlerInitMultipartUpload(ctx echo.Context) error {
	var initupload core2.InitMultipartUpload
	err := ctx.Bind(&initupload)
	if err != nil {
		logx.GetLogger("study").Errorf("HandlerInitMultipartUpload|Bind err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	initupload.WithBucket(cs.bucket)

	// 初始化上传文件
	uploadid, err := initupload.InitUpload(ctx, cs.s3Client)
	if err != nil {
		logx.GetLogger("study").Errorf("HandlerInitMultipartUpload|InitUpload err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "InitUpload Error",
		})
	}

	initupload.WithUploadId(uploadid)
	marshal, err := json.Marshal(&initupload)
	if err != nil {
		logx.GetLogger("study").Errorf("HandlerInitMultipartUpload|Marshal err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Marshal Error",
		})
	}

	// 添加redis中的init文件
	_, err = cs.redis.Set(ctx.Request().Context(), fmt.Sprintf(core2.RedisInitIndexKey, initupload.Fid), string(marshal), 0).Result()
	if err != nil {
		logx.GetLogger("study").Errorf("HandlerInitMultipartUpload|Set err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Redis Set Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, initupload)
}

func (cs *CosServer) HandlerMultiUpload(ctx echo.Context) error {
	fid := ctx.Param("fid")
	if len(fid) <= 0 {
		logx.GetLogger("study").Errorf("HandlerMultiUpload|fid is empty")
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	partidStr := ctx.QueryParam("partid")
	if len(partidStr) <= 0 {
		logx.GetLogger("study").Errorf("HandlerMultiUpload|partid is empty")
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	partid, _ := strconv.ParseInt(partidStr, 10, 64)

	init, err := core2.QueryIndexToInit(ctx, cs.redis, fid)
	if err != nil {
		logx.GetLogger("study").Errorf("HandlerMultiUpload|QueryIndexToInit err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "QueryIndexToInit Error",
		})
	}

	ETag, err := init.MultipartUpload(ctx, int(partid), cs.s3Client)
	if err != nil {
		logx.GetLogger("study").Errorf("HandlerMultiUpload|MultipartUpload err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "MultipartUpload Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, echo.Map{
		"partid": partid,
		"etag":   ETag,
	})
}

type EndPart struct {
	PartId int32  `json:"part_id"`
	ETag   string `json:"etag"`
}

func (cs *CosServer) CompleteUpload(ctx echo.Context) error {

	fid := ctx.Param("fid")
	if len(fid) <= 0 {
		logx.GetLogger("study").Errorf("CompleteUpload|fid is empty")
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	var endparts []EndPart
	err := ctx.Bind(&endparts)
	if err != nil {
		logx.GetLogger("study").Errorf("CompleteUpload|Bind err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	// redis中获取上传文件信息
	result, err := cs.redis.Get(ctx.Request().Context(), fmt.Sprintf(core2.RedisInitIndexKey, fid)).Result()
	if err != nil {
		logx.GetLogger("study").Errorf("CompleteUpload|Get err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Redis Get Error",
		})
	}

	var init core2.InitMultipartUpload
	err = json.Unmarshal([]byte(result), &init)
	if err != nil {
		logx.GetLogger("study").Errorf("CompleteUpload|Unmarshal err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "json Unmarshal Error",
		})
	}

	completeParts := make([]types.CompletedPart, len(endparts))
	for i, endpart := range endparts {
		completeParts[i] = types.CompletedPart{
			ETag:       &endpart.ETag,
			PartNumber: &endpart.PartId,
		}
	}

	_, err = cs.s3Client.CompleteMultipartUpload(ctx.Request().Context(), &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(cs.bucket),
		Key:      aws.String(init.GetFilePath()),
		UploadId: aws.String(init.UploadId),
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: completeParts,
		},
	})
	if err != nil {
		logx.GetLogger("study").Errorf("CompleteUpload|CompleteMultipartUpload err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "CompleteMult Error",
		})
	}
	// 修改prepare文件为index
	prepareKey := fmt.Sprintf(core2.RedisPrepareIndexKey, init.Fid)
	index := fmt.Sprintf(core2.RedisIndexKey, init.Fid)
	err = cs.redis.RenameNX(ctx.Request().Context(), prepareKey, index).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("CompleteUpload|RenameNX err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "RenameNX Error",
		})
	}
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, echo.Map{
		"msg": "CompleteUpload Success",
	})
}

func (cs *CosServer) HandlerAbortUpload(ctx echo.Context) error {
	fid := ctx.Param("fid")
	if len(fid) == 0 {
		logx.GetLogger("study").Errorf("HandlerAbortUpload|fid is empty")
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	// 查询init文件
	init, err := core2.QueryIndexToInit(ctx, cs.redis, fid)
	if err != nil {
		logx.GetLogger("study").Errorf("HandlerAbortUpload|QueryIndexToInit err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "QueryIndexToInit Error",
		})
	}

	_, err = cs.s3Client.AbortMultipartUpload(ctx.Request().Context(), &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(cs.bucket),
		Key:      aws.String(init.GetFilePath()),
		UploadId: aws.String(init.UploadId),
	})
	if err != nil {
		logx.GetLogger("study").Errorf("HandlerAbortUpload|AbortMultipartUpload err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "AbortMultipartUpload Error",
		})
	}

	// 删除redis中的prepare文件和init文件
	prepareKey := fmt.Sprintf(core2.RedisPrepareIndexKey, fid)
	initKey := fmt.Sprintf(core2.RedisInitIndexKey, fid)
	err = cs.redis.Del(ctx.Request().Context(), prepareKey, initKey).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("HandlerAbortUpload|Del err|%s|%s|%v", prepareKey, initKey, err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Redis Del Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, echo.Map{
		"msg": "AbortUpload Success",
	})
}

func (cs *CosServer) checkAndCreateBucket() error {
	for _, bucket := range config.GloableConfig.S3.Bucket {
		_, err := cs.s3Client.HeadBucket(context.TODO(), &s3.HeadBucketInput{
			Bucket: bucket,
		})
		if err != nil {
			logx.GetLogger("study").Errorf("CheckAndCreateBucket|HeadBucket err:%v", err)

			// 创建bucket
			_, err = cs.s3Client.CreateBucket(context.TODO(), &s3.CreateBucketInput{
				Bucket: bucket,
			})
			if err != nil {
				logx.GetLogger("study").Errorf("CheckAndCreateBucket|CreateBucket err:%v", err)
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
				logx.GetLogger("study").Errorf("CheckAndCreateBucket|PutBucketPolicy err:%v", err)
				return err
			}
		}
	}
	return nil
}
