package service

import (
	"context"
	"cos/api/config"
	"cos/api/core"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/dzjyyds666/opensource/common"
	"github.com/dzjyyds666/opensource/httpx"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/labstack/echo"
	"github.com/redis/go-redis/v9"
)

type CosService struct {
	bucket    string
	tmpPart   string // 临时文件存放
	cosServer *core.CosFileServer
}

func NewCosServer(ctx context.Context, ds *redis.Client, s3Client *s3.Client) (*CosService, error) {
	// 初始化s3客户端

	cosServer := &CosService{
		cosServer: core.NewCosFileServer(ctx, ds, s3Client),
		bucket:    aws.ToString(config.GloableConfig.S3.Bucket[0]),
		tmpPart:   "/tmp/study",
	}

	err := cosServer.checkAndCreateBucket()
	if err != nil {
		logx.GetLogger("study").Errorf("NewCosServer|CheckAndCreateBucket|err:%v", err)
		return nil, err
	}
	return cosServer, nil
}

// 需要传入文件的名称，md5，长度，文件夹id
func (cs *CosService) HandleApplyUpload(ctx echo.Context) error {

	decoder := json.NewDecoder(ctx.Request().Body)
	var cosFile core.CosFile
	err := decoder.Decode(&cosFile)
	if err != nil {
		logx.GetLogger("study").Infof("HandleApplyUpload|Params bind Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	logx.GetLogger("study").Infof("HandleApplyUpload|Params bind Success|%s", common.ToStringWithoutError(cosFile))

	// uuid生成fid
	fid := core.GenerateFid()
	cosFile.WithFid("fi_" + fid)

	err = cs.cosServer.ApplyUpload(ctx.Request().Context(), &cosFile)
	if err != nil {
		logx.GetLogger("study").Errorf("HandleApplyUpload|CreatePrepareIndex|err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Create File PrepareInfo Error",
		})
	}
	logx.GetLogger("study").Infof("HandleApplyUpload|CreatePrepareIndex|Succ|%s", common.ToStringWithoutError(cosFile))
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, cosFile)
}

func (cs *CosService) HandleInitUploadVideo(ctx echo.Context) error {
	var init core.InitMultipartUpload
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&init); err != nil {
		logx.GetLogger("study").Errorf("HandleInitUploadVideo|Decode err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	err := cs.cosServer.InitUploadVideo(ctx.Request().Context(), &init)
	if err != nil {
		logx.GetLogger("study").Errorf("HandleInitUploadVideo|InitUpload err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "InitUpload Error",
		})
	}
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, init)
}

func (cs *CosService) HandleUploadVideoPart(ctx echo.Context) error {
	fid := ctx.QueryParam("fid")
	partId := ctx.QueryParam("partId")
	if len(fid) <= 0 || len(partId) <= 0 {
		logx.GetLogger("study").Errorf("HandleUploadPartToTmp|fid or partId is empty")
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		logx.GetLogger("study").Errorf("HandleUploadPartToTmp|Decode err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	open, err := file.Open()
	if err != nil {
		logx.GetLogger("study").Errorf("HandleUploadPartToTmp|Open err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	defer open.Close()

	err = cs.cosServer.UploadVideoPart(ctx.Request().Context(), fid, partId, open)
	if err != nil {
		logx.GetLogger("study").Errorf("HandleUploadPartToTmp|UploadPart err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "UploadPart Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, echo.Map{
		"msg": "UploadPart Success",
	})
}

func (cs *CosService) HandleCompleteUploadVideo(ctx echo.Context) error {

	fid := ctx.Param("fid")
	if len(fid) <= 0 {
		logx.GetLogger("study").Errorf("HandleCompleteUploadVideo|fid is empty")
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	err := cs.cosServer.CompleteUploadVideo(ctx.Request().Context(), fid)
	if err != nil {
		logx.GetLogger("study").Errorf("HandleCompleteUploadVideo|CompleteUploadVideo err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "CompleteUploadVideo Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, echo.Map{
		"msg": "Merge File Success",
	})
}

func (cs *CosService) HandleSingleUpload(ctx echo.Context) error {
	fid := ctx.Param("fid")
	file, err := ctx.FormFile("file")
	if err != nil || len(fid) <= 0 {
		logx.GetLogger("study").Errorf("HandleSingleUpload|FormFile err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	open, err := file.Open()
	if err != nil {
		logx.GetLogger("study").Errorf("HandleSingleUpload|Open err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	defer open.Close()

	info, err := cs.cosServer.SingleUpload(ctx.Request().Context(), cs.bucket, fid, open)
	if err != nil {
		logx.GetLogger("study").Errorf("HandleSingleUpload|SingleUpload err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "SingleUpload Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, info)
}

func (cs *CosService) HandleInitMultipartUpload(ctx echo.Context) error {
	var initupload core.InitMultipartUpload
	decoder := json.NewDecoder(ctx.Request().Body)
	err := decoder.Decode(&initupload)
	if err != nil {
		logx.GetLogger("study").Errorf("HandleInitMultipartUpload|Decode err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	err = cs.cosServer.InitUpload(ctx.Request().Context(), cs.bucket, &initupload)
	if err != nil {
		logx.GetLogger("study").Errorf("HandleInitMultipartUpload|InitMultipartUpload err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "InitMultipartUpload Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, initupload)
}

func (cs *CosService) HandleUploadPart(ctx echo.Context) error {
	fid := ctx.QueryParam("fid")
	partId := ctx.QueryParam("partId")
	if len(fid) <= 0 || len(partId) <= 0 {
		logx.GetLogger("study").Errorf("HandleUploadPart|fid or partId is empty")
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		logx.GetLogger("study").Errorf("HandleUploadPart|Decode err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	open, err := file.Open()
	if err != nil {
		logx.GetLogger("study").Errorf("HandleUploadPart|Open err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	defer open.Close()

	err = cs.cosServer.UploadPart(ctx.Request().Context(), cs.bucket, fid, partId, open)
	if err != nil {
		logx.GetLogger("study").Errorf("HandleUploadPart|UploadPart err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "UploadPart Error",
		})
	}
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, echo.Map{
		"msg": "UploadPart Success",
	})
}

type EndPart struct {
	PartId int32  `json:"part_id"`
	ETag   string `json:"etag"`
}

func (cs *CosService) CompleteUpload(ctx echo.Context) error {

	fid := ctx.Param("fid")
	if len(fid) <= 0 {
		logx.GetLogger("study").Errorf("CompleteUpload|fid is empty")
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	var endparts []EndPart
	decoder := json.NewDecoder(ctx.Request().Body)
	err := decoder.Decode(&endparts)
	if err != nil {
		logx.GetLogger("study").Errorf("CompleteUpload|Bind err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	completeParts := make([]types.CompletedPart, len(endparts)-1)
	for i, endpart := range endparts {
		completeParts[i] = types.CompletedPart{
			ETag:       &endpart.ETag,
			PartNumber: &endpart.PartId,
		}
	}

	err = cs.cosServer.CompleteMultUpload(ctx.Request().Context(), cs.bucket, fid, completeParts)
	if err != nil {
		logx.GetLogger("study").Errorf("CompleteUpload|CompleteMultipartUpload err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "CompleteMult Error",
		})
	}
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, echo.Map{
		"msg": "CompleteUpload Success",
	})
}

func (cs *CosService) HandleAbortUpload(ctx echo.Context) error {
	fid := ctx.Param("fid")
	if len(fid) == 0 {
		logx.GetLogger("study").Errorf("HandleAbortUpload|fid is empty")
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	err := cs.cosServer.AbortUpload(ctx.Request().Context(), cs.bucket, fid)
	if err != nil {
		logx.GetLogger("study").Errorf("HandleAbortUpload|AbortUpload err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "AbortUpload Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, echo.Map{
		"msg": "AbortUpload Success",
	})
}

func (cs *CosService) checkAndCreateBucket() error {
	err := cs.cosServer.CheckBucketExist()
	if err != nil {
		logx.GetLogger("study").Errorf("CheckAndCreateBucket|CheckBucketExist err:%v", err)
		panic(err)
	}
	return nil
}
