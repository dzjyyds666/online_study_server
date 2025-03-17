package core

import (
	"context"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/redis/go-redis/v9"
	"time"
)

type ClassResource struct {
	Fid      *string `json:"fid"`
	FileName *string `json:"file_name"`
	FileType *string `json:"file_type"`
	FileSize *int64  `json:"file_size"`
	UploadTs *int64  `json:"upload_ts"`
}

func (r *ClassResource) WithFid(id string) *ClassResource {
	r.Fid = &id
	return r
}

func (r *ClassResource) WithFileName(name string) *ClassResource {
	r.FileName = &name
	return r
}

func (r *ClassResource) WithFileType(type_ string) *ClassResource {
	r.FileType = &type_
	return r
}

func (r *ClassResource) WithFileSize(size int64) *ClassResource {
	r.FileSize = &size
	return r
}

func (r *ClassResource) WithUploadTs(ts int64) *ClassResource {
	r.UploadTs = &ts
	return r
}

func (cr *ClassResource) CreateResource(ctx context.Context, chid string, ds *redis.Client) error {
	// 添加资源的索引
	resourceKey := BuildChapterResourceIndexKey(*cr.Fid)
	_, err := ds.Set(ctx, resourceKey, cr, 0).Result()
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("CreateResource|Set Resource Error|%v", err)
		return err
	}

	// 把资源添加到章节下面
	chapterKey := BuildChapterResourceKey(chid)
	_, err = ds.ZAdd(ctx, chapterKey, redis.Z{
		Member: cr.Fid,
		Score:  float64(time.Now().Unix()),
	}).Result()

	if err != nil {
		logx.GetLogger("OS_Server").Errorf("CreateResource|Add Resource Error|%v", err)
		return err
	}

	logx.GetLogger("OS_Server").Infof("CreateResource|Create Resource Success|%v", cr)
	return nil
}

func (cr *ClassResource) DeleteResource(ctx context.Context, ds *redis.Client) {
}
