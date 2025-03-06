package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/labstack/echo"
	"github.com/redis/go-redis/v9"
	"path"
)

type InitMultipartUpload struct {
	FileSize    int64  `json:"file_size,omitempty"`
	FileType    string `json:"file_type,omitempty"`
	DirectoryId string `json:"directory_id,omitempty"`
	Bucket      string `json:"bucket,omitempty"`
	Fid         string `json:"fid,omitempty"`
	FileNmae    string `json:"file_name,omitempty"`
	PartBytes   int64  `json:"part_bytes,omitempty"`
	TotalParts  int64  `json:"total_parts,omitempty"`
	LastParts   int64  `json:"last_parts,omitempty"`
	UploadId    string `json:"upload_id,omitempty"`
}

func (imu *InitMultipartUpload) GetFilePath() string {
	return fmt.Sprintf("%s/%s%s", imu.DirectoryId, imu.Fid, path.Ext(imu.FileNmae))
}

func (imu *InitMultipartUpload) WithFid(fid string) *InitMultipartUpload {
	imu.Fid = fid
	return imu
}

func (imu *InitMultipartUpload) WithUploadId(uploadId string) *InitMultipartUpload {
	imu.UploadId = uploadId
	return imu
}

func (imu *InitMultipartUpload) WithFileSize(fileSize int64) *InitMultipartUpload {
	imu.FileSize = fileSize
	return imu
}

func (imu *InitMultipartUpload) WithFileType(fileType string) *InitMultipartUpload {
	imu.FileType = fileType
	return imu
}

func (imu *InitMultipartUpload) WithDirectoryId(directoryId string) *InitMultipartUpload {
	imu.DirectoryId = directoryId
	return imu
}

func (imu *InitMultipartUpload) WithBucket(bucket string) *InitMultipartUpload {
	imu.Bucket = bucket
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

func (imu *InitMultipartUpload) InitUpload(ctx echo.Context, client *s3.Client) (string, error) {
	upload, err := client.CreateMultipartUpload(ctx.Request().Context(), &s3.CreateMultipartUploadInput{
		Bucket: aws.String(imu.Bucket),
		Key:    aws.String(imu.GetFilePath()),
	})
	if err != nil {
		return "", err
	}
	uploadid := upload.UploadId
	return *uploadid, nil
}

func QueryIndexToInit(ctx echo.Context, rs *redis.Client, fid string) (*InitMultipartUpload, error) {
	// 从redis中获取到prepare文件
	if len(fid) <= 0 {
		return nil, ErrFidNotExits
	}

	key := fmt.Sprintf(RedisPrepareIndexKey, fid)
	result, err := rs.Get(ctx.Request().Context(), key).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		logx.GetLogger("OS_Server").Errorf("QueryPrepareIndex|Get Error|%v", err)
		return nil, err
	}

	if errors.Is(err, redis.Nil) {
		logx.GetLogger("OS_Server").Infof("QueryPrepareIndex|Prepare Index Not Exits|%v", err)
		return nil, err
	}

	var init InitMultipartUpload
	err = json.Unmarshal([]byte(result), &init)
	if err != nil {
		return nil, err
	}

	return &init, nil
}

func (cf *InitMultipartUpload) MultipartUpload(ctx echo.Context, partId int, client *s3.Client) (string, error) {

	input := &s3.UploadPartInput{
		Bucket:     aws.String(cf.Bucket),
		Key:        aws.String(cf.GetFilePath()),
		UploadId:   aws.String(cf.UploadId),
		PartNumber: aws.Int32(int32(partId)),
		Body:       ctx.Request().Body,
	}

	part, err := client.UploadPart(ctx.Request().Context(), input)
	if err != nil {
		return "", err
	}

	return aws.ToString(part.ETag), nil
}
