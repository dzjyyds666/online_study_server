package core

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/google/uuid"
	"github.com/labstack/echo"
	"github.com/redis/go-redis/v9"
	"io"
)

type CosFile struct {
	FileName    *string `json:"file_name,omitempty"`
	Fid         *string `json:"fid,omitempty"`
	FileMD5     *string `json:"file_md5,omitempty"`
	FileSize    *int64  `json:"file_size,omitempty"`
	FileType    *string `json:"file_type,omitempty"`
	DirectoryId *string `json:"directory_id,omitempty"`
	IsMultiPart *bool   `json:"is_multi_part,omitempty"`

	r io.Reader
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

func (cf *CosFile) WithReader(r io.Reader) *CosFile {
	cf.r = r
	return cf
}

var ErrPrepareIndexExits = fmt.Errorf("CraetePrepareIndex Exits")

func (cf *CosFile) CraetePrepareIndex(ctx echo.Context, redis *redis.Client) error {
	marshal, err := json.Marshal(cf)
	if err != nil {
		logx.GetLogger("OS_Server").Infof("CraetePrepareIndex|Marshal Error|%v", err)
		return err
	}

	exists, err := redis.SetNX(ctx.Request().Context(), fmt.Sprintf(RedisPrepareIndexKey, *cf.Fid), marshal, 0).Result()
	if err != nil {
		logx.GetLogger("OS_Server").Infof("CraetePrepareIndex|SetNX Error|%v", err)
		return err
	}
	if !exists {
		logx.GetLogger("OS_Server").Errorf("CraetePrepareIndex|CraetePrepareIndex Exits")
		return ErrPrepareIndexExits
	}

	return nil
}

func (cf *CosFile) CraeteIndex(ctx echo.Context, redis *redis.Client) error {
	marshal, err := json.Marshal(cf)
	if err != nil {
		logx.GetLogger("OS_Server").Infof("CraeteIndex|Marshal Error|%v", err)
		return err
	}

	// 从redis中删除prepare文件
	_, err = redis.Del(ctx.Request().Context(), fmt.Sprintf(RedisPrepareIndexKey, *cf.Fid)).Result()
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("CraeteIndex|Del Error|%v", err)
		return err
	}

	// 插入index文件到redis中
	_, err = redis.SetNX(ctx.Request().Context(), fmt.Sprintf(RedisIndexKey, *cf.Fid), marshal, 0).Result()
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("CraeteIndex|Set Error|%v", err)
		return err
	}
	return nil
}

func (cf *CosFile) IsMatch(cos *CosFile) bool {
	return *cf.FileName == *cos.FileName && *cf.FileMD5 == *cos.FileMD5 && *cf.FileSize == *cos.FileSize
}

func (cf *CosFile) GetFilePath() string {
	return fmt.Sprintf("/%s/%s", *cf.DirectoryId, *cf.Fid)
}

func (cf *CosFile) PutObject(ctx echo.Context, client *s3.Client, bucket *string) error {
	key := cf.GetFilePath()

	_, err := client.PutObject(ctx.Request().Context(), &s3.PutObjectInput{
		Body:   cf.r,
		Bucket: bucket,
		Key:    aws.String(key),
		//ContentMD5:    cf.FileMD5,
		ContentLength: cf.FileSize,
		ContentType:   cf.FileType,
	})
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("PutObject Error|%v", err)
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

func GenerateFid() string {
	u := uuid.New()
	return u.String()
}

var ErrFidNotExits = fmt.Errorf("Fid Not Exits")

func QueryPrepareIndex(ctx echo.Context, rs *redis.Client, fid string) (*CosFile, error) {
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

	var prepareFile CosFile

	if err := json.Unmarshal([]byte(result), &prepareFile); err != nil {
		logx.GetLogger("OS_Server").Errorf("QueryPrepareIndex|Unmarshal Error|%v", err)
		return nil, err
	}

	return &prepareFile, nil
}

func QueryIndex(ctx echo.Context, rs *redis.Client, fid string) (*CosFile, error) {
	// 从redis中获取到prepare文件
	if len(fid) <= 0 {
		return nil, ErrFidNotExits
	}

	key := fmt.Sprintf(RedisIndexKey, fid)
	result, err := rs.Get(ctx.Request().Context(), key).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		logx.GetLogger("OS_Server").Errorf("QueryPrepareIndex|Get Error|%v", err)
		return nil, err
	}

	if errors.Is(err, redis.Nil) {
		logx.GetLogger("OS_Server").Infof("QueryPrepareIndex|Prepare Index Not Exits|%v", err)
		return nil, err
	}

	var indexFile CosFile

	if err := json.Unmarshal([]byte(result), &indexFile); err != nil {
		logx.GetLogger("OS_Server").Errorf("QueryPrepareIndex|Unmarshal Error|%v", err)
		return nil, err
	}

	return &indexFile, nil
}
