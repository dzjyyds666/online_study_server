package core

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/labstack/echo"
	"github.com/redis/go-redis/v9"
	"io"
	"os"
	"path"
)

type CosFile struct {
	FileName    *string `json:"file_name,omitempty"`
	Fid         *string `json:"fid,omitempty"`
	FileMD5     *string `json:"file_md5,omitempty"`
	FileSize    *int64  `json:"file_size,omitempty"`
	FileType    *string `json:"file_type,omitempty"`
	DirectoryId *string `json:"directory_id,omitempty"`
}

func (cf *CosFile) MergeFilePath() string {
	return fmt.Sprintf("/%s/%s/%s%s", *cf.DirectoryId, *cf.Fid, *cf.Fid, path.Ext(*cf.FileName))
}

func (cf *CosFile) WithFileName(fileName string) *CosFile {
	cf.FileName = aws.String(fileName)
	return cf
}

func (cf *CosFile) WithFid(fid string) *CosFile {
	cf.Fid = aws.String(fid)
	return cf
}

func (cf *CosFile) WithFileMD5(fileMD5 string) *CosFile {
	cf.FileMD5 = aws.String(fileMD5)
	return cf
}

func (cf *CosFile) WithFileSize(fileSize int64) *CosFile {
	cf.FileSize = aws.Int64(fileSize)
	return cf
}

func (cf *CosFile) WithFileType(fileType string) *CosFile {
	cf.FileType = aws.String(fileType)
	return cf
}

func (cf *CosFile) WithDirectoryId(directoryId string) *CosFile {
	cf.DirectoryId = aws.String(directoryId)
	return cf
}

func (cf *CosFile) Marshal() string {
	marshal, _ := json.Marshal(cf)
	return string(marshal)
}

var ErrPrepareIndexExits = fmt.Errorf("CreatePrepareIndex Exits")

func (cf *CosFile) CreatePrepareIndex(ctx echo.Context, redis *redis.Client) error {
	marshal, err := json.Marshal(cf)
	if err != nil {
		logx.GetLogger("study").Infof("CreatePrepareIndex|Marshal Error|%v", err)
		return err
	}

	exists, err := redis.SetNX(ctx.Request().Context(), fmt.Sprintf(RedisPrepareInfoKey, *cf.Fid), marshal, 0).Result()
	if err != nil {
		logx.GetLogger("study").Infof("CreatePrepareIndex|SetNX Error|%v", err)
		return err
	}
	if !exists {
		logx.GetLogger("study").Errorf("CreatePrepareIndex|CreatePrepareIndex Exits")
		return ErrPrepareIndexExits
	}

	return nil
}

func (cf *CosFile) CreateIndex(ctx echo.Context, redis *redis.Client) error {
	marshal, err := json.Marshal(cf)
	if err != nil {
		logx.GetLogger("study").Infof("CreateIndex|Marshal Error|%v", err)
		return err
	}

	// 从redis中删除prepare文件
	_, err = redis.Del(ctx.Request().Context(), fmt.Sprintf(RedisPrepareInfoKey, *cf.Fid)).Result()
	if err != nil {
		logx.GetLogger("study").Errorf("CreateIndex|Del Error|%v", err)
		return err
	}

	// 插入index文件到redis中
	_, err = redis.Set(ctx.Request().Context(), fmt.Sprintf(RedisInfoKey, *cf.Fid), marshal, 0).Result()
	if err != nil {
		logx.GetLogger("study").Errorf("CreateIndex|Set Error|%v", err)
		return err
	}
	return nil
}

func (cf *CosFile) IsMatch(cos *CosFile) bool {
	return *cf.FileName == *cos.FileName && *cf.FileMD5 == *cos.FileMD5 && *cf.FileSize == *cos.FileSize
}

func (cf *CosFile) UploadSingleFile(ctx echo.Context, client *s3.Client, bucket *string, ds *redis.Client) error {
	// 先上传文件到minio
	err := cf.PutObject(ctx, client, bucket)
	if err != nil {
		logx.GetLogger("study").Infof("UploadSingleFile|PutObject Error|%v", err)
		return err
	}

	// 插入源文件的indexInfo
	err = cf.CreateIndex(ctx, ds)
	if err != nil {
		logx.GetLogger("study").Errorf("UploadSingleFile|CreateIndex Error|%v", err)
		return err
	}
	logx.GetLogger("study").Infof("UploadSingleFile|UploadSingleFile Success")
	return err
}

func (cf *CosFile) PutObject(ctx echo.Context, client *s3.Client, bucket *string) error {
	key := cf.MergeFilePath()

	logx.GetLogger("study").Infof("PutObject|%v", *cf.FileType)

	_, err := client.PutObject(ctx.Request().Context(), &s3.PutObjectInput{
		Body:   cf.r,
		Bucket: bucket,
		Key:    aws.String(key),
		//ContentMD5:    cf.FileMD5,
		ContentLength: cf.FileSize,
		ContentType:   cf.FileType,
	})
	if err != nil {
		logx.GetLogger("study").Errorf("PutObject Error|%v", err)
		return err
	}

	return nil
}

// todo 计算文件的md5
func CalculateMD5(reader io.Reader) (string, error) {
	buffer := make([]byte, 1024)
	md5Hash := md5.New()
	for {
		n, err := reader.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		md5Hash.Write(buffer[:n])
	}
	return hex.EncodeToString(md5Hash.Sum(nil)), nil
}

func GetFileType(reader io.Reader) (string, error) {
	buffer := make([]byte, 512)
	n, err := reader.Read(buffer)
	if err != nil {
		return "", err
	}

	detect := mimetype.Detect(buffer[:n])

	return detect.String(), nil
}

func GenerateFid() string {
	u := uuid.New()
	return u.String()
}

var ErrFidNotExits = fmt.Errorf("Fid Not Exits")
