package core

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
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
	DirectoryId *string `json:"directory_id,omitempty"`
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

func (cf *CosFile) WithDirectoryId(directoryId string) *CosFile {
	cf.DirectoryId = aws.String(directoryId)
	return cf
}

var ErrPrepareIndexExits = fmt.Errorf("PrepareIndex Exits")

func (cf *CosFile) PrepareIndex(ctx echo.Context, redis *redis.Client) error {
	marshal, err := json.Marshal(cf)
	if err != nil {
		logx.GetLogger("OS_Server").Infof("PrepareIndex|Marshal Error|%v", err)
		return err
	}

	exists, err := redis.SetNX(ctx.Request().Context(), fmt.Sprintf(RedisPrepareIndexKey, *cf.DirectoryId, *cf.Fid), marshal, 0).Result()
	if err != nil {
		logx.GetLogger("OS_Server").Infof("PrepareIndex|SetNX Error|%v", err)
		return err
	}
	if !exists {
		logx.GetLogger("OS_Server").Errorf("PrepareIndex|PrepareIndex Exits")
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
	_, err = redis.Del(ctx.Request().Context(), fmt.Sprintf(RedisPrepareIndexKey, *cf.DirectoryId, *cf.Fid)).Result()
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("CraeteIndex|Del Error|%v", err)
		return err
	}

	// 插入index文件到redis中
	_, err = redis.SetNX(ctx.Request().Context(), fmt.Sprintf(RedisIndexKey, *cf.DirectoryId, *cf.Fid), marshal, 0).Result()
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
	return fmt.Sprintf("%s/%s", *cf.DirectoryId, *cf.Fid)
}

func CalculateMD5(r io.Reader) (string, error) {
	hash := md5.New()

	_, err := io.Copy(hash, r)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func GenerateFid() string {
	u := uuid.New()
	return u.String()
}

func QueryPrepareIndex(ctx echo.Context, redis *redis.Client, file *CosFile) (*CosFile, error) {
	return nil, nil
}
