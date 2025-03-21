package core

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dzjyyds666/opensource/common"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/redis/go-redis/v9"
)

const (
	RedisResourceKey = "class:resource:%s" // fid
)

func BuildResourceKey(fid string) string {
	return fmt.Sprintf(RedisResourceKey, fid)
}

type Resource struct {
	Fid          string `json:"fid"`
	FileName     string `json:"file_name"`
	FileSize     int64  `json:"file_size"`
	FileType     string `json:"file_type"`
	Published    bool   `json:"published"`
	Downloadable bool   `json:"downloadable"` // 是否可以下载
}

func (r *Resource) WithFid(fid string) *Resource {
	r.Fid = fid
	return r
}

func (r *Resource) WithFileName(fileName string) *Resource {
	r.FileName = fileName
	return r
}

func (r *Resource) WithFileType(fileType string) *Resource {
	r.FileType = fileType
	return r
}

func (r *Resource) WithFileSize(fileSize int64) *Resource {
	r.FileSize = fileSize
	return r
}

func (r *Resource) WithPublished(published bool) *Resource {
	r.Published = published
	return r
}

func (r *Resource) WithDownloadable(downloadable bool) *Resource {
	r.Downloadable = downloadable
	return r
}

func (r *Resource) IsPublished() bool {
	return r.Published == true
}

func (r *Resource) IsDownloadable() bool {
	return r.Downloadable == true
}

func (r *Resource) UpdatePublishResource(ctx context.Context, ds *redis.Client) error {
	result, err := ds.Get(ctx, BuildResourceKey(r.Fid)).Result()
	if err != nil {
		logx.GetLogger("study").Errorf("PublishResource|Get Resource Error|%v", err)
		return err
	}

	var resource Resource
	if err := json.Unmarshal([]byte(result), &resource); err != nil {
		logx.GetLogger("study").Errorf("PublishResource|Unmarshal Resource Info Error|%v", err)
		return err
	}

	resource.WithPublished(r.Published)

	marshal, err := json.Marshal(&resource)
	if nil != err {
		logx.GetLogger("study").Errorf("PublishResource|Marshal Resource Error|%v", err)
		return err
	}

	if err := ds.Set(ctx, BuildResourceKey(r.Fid), marshal, 0).Err(); err != nil {
		logx.GetLogger("study").Errorf("PublishResource|Set Resource Error|%v", err)
		return err
	}

	logx.GetLogger("study").Infof("PublishResource|Update Resource Published Statuc Success|%v", common.ToStringWithoutError(resource))
	return nil
}
func (r *Resource) CreateUploadResource(ctx context.Context, ds *redis.Client) error {
	raw, err := json.Marshal(r)
	if nil != err {
		logx.GetLogger("study").Errorf("CreateUploadResource|Marshal Resource Error|%v", err)
		return err
	}
	key := BuildResourceInfo(r.Fid)
	err = ds.Set(ctx, key, raw, 0).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("CreateUploadResource|Set Resource Error|%v", err)
		return err
	}
	return nil
}

func (r *Resource) DeleteUploadResource(ctx context.Context, ds *redis.Client) error {
	err := ds.Del(ctx, BuildResourceKey(r.Fid)).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("DeleteUploadResource|Delete Resource Error|%v", err)
		return err
	}
	return nil
}
