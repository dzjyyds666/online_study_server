package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/labstack/echo"
	"github.com/redis/go-redis/v9"
)

var InitStatus = struct {
	Ing  string `json:"ing"`
	Done string `json:"done"`
	Fail string `json:"fail"`
}{
	Ing:  "ing",
	Done: "done",
	Fail: "fail",
}

type InitMultipartUpload struct {
	Fid        string `json:"fid,omitempty"`
	PartBytes  int64  `json:"part_bytes,omitempty"`
	TotalParts int64  `json:"total_parts,omitempty"`
	LastParts  int64  `json:"last_parts,omitempty"`
	UploadId   string `json:"upload_id,omitempty"`
	Status     string `json:"status,omitempty"` // 状态
}

func (imu *InitMultipartUpload) WithFid(fid string) *InitMultipartUpload {
	imu.Fid = fid
	return imu
}

func (imu *InitMultipartUpload) WithUploadId(uploadId string) *InitMultipartUpload {
	imu.UploadId = uploadId
	return imu
}

func (imu *InitMultipartUpload) WithPartBytes(partBytes int64) *InitMultipartUpload {
	imu.PartBytes = partBytes
	return imu
}

func (imu *InitMultipartUpload) WithTotalParts(totalParts int64) *InitMultipartUpload {
	imu.TotalParts = totalParts
	return imu
}

func (imu *InitMultipartUpload) WithLastParts(lastParts int64) *InitMultipartUpload {
	imu.LastParts = lastParts
	return imu
}

func (imu *InitMultipartUpload) Marshal() string {
	data, _ := json.Marshal(imu)
	return string(data)
}

func (imu *InitMultipartUpload) WithStatus(status string) *InitMultipartUpload {
	imu.Status = status
	return imu
}

func (imu *InitMultipartUpload) InitUpload(ctx echo.Context, bucket, objectKey string, client *s3.Client) (string, error) {

	upload, err := client.CreateMultipartUpload(ctx.Request().Context(), &s3.CreateMultipartUploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return "", err
	}
	uploadid := upload.UploadId
	return *uploadid, nil
}

func (imu *InitMultipartUpload) QueryIndexToInit(ctx context.Context, rs *redis.Client) (*InitMultipartUpload, error) {
	// 从redis中获取到prepare文件
	if len(imu.Fid) <= 0 {
		return nil, ErrFidNotExits
	}

	key := fmt.Sprintf(RedisPrepareInfoKey, imu.Fid)
	result, err := rs.Get(ctx, key).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		logx.GetLogger("study").Errorf("QueryPrepareIndex|Get Error|%v", err)
		return nil, err
	}

	if errors.Is(err, redis.Nil) {
		logx.GetLogger("study").Infof("QueryPrepareIndex|Prepare Index Not Exits|%v", err)
		return nil, err
	}

	var init InitMultipartUpload
	err = json.Unmarshal([]byte(result), &init)
	if err != nil {
		return nil, err
	}

	return &init, nil
}

func (imu *InitMultipartUpload) CreateInitIndex(ctx context.Context, rs *redis.Client) error {
	key := fmt.Sprintf(RedisInitInfoKey, imu.Fid)
	value, err := json.Marshal(imu)
	if err != nil {
		return err
	}
	return rs.Set(ctx, key, value, 0).Err()
}
