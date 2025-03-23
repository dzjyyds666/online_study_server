package core

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dzjyyds666/opensource/common"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/redis/go-redis/v9"
	"math"
	"strconv"
	"time"
)

const (
	RedisResourceKey = "class:resource:%s" // fid
)

func BuildResourceKey(fid string) string {
	return fmt.Sprintf(RedisResourceKey, fid)
}

type Resource struct {
	Fid          string `json:"fid,omitempty"`
	FileName     string `json:"file_name,omitempty"`
	FileSize     int64  `json:"file_size,omitempty"`
	FileType     string `json:"file_type,omitempty"`
	FileMd5      string `json:"file_md5,omitempty"`
	Published    bool   `json:"published"`
	Downloadable bool   `json:"downloadable"` // 是否可以下载
	Chid         string `json:"chid,omitempty"`
}

func (r *Resource) WithFid(fid string) *Resource {
	r.Fid = fid
	return r
}

func (r *Resource) WithChid(chid string) *Resource {
	r.Chid = chid
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

func (r *Resource) UpdatePublishResource(ctx context.Context, ds *redis.Client) (*Resource, error) {
	infoKey := BuildResourceInfo(r.Fid)
	result, err := ds.Get(ctx, infoKey).Result()
	if err != nil {
		logx.GetLogger("study").Errorf("PublishResource|Get Resource Error|%v", err)
		return nil, err
	}

	var resource Resource
	if err := json.Unmarshal([]byte(result), &resource); err != nil {
		logx.GetLogger("study").Errorf("PublishResource|Unmarshal Resource Info Error|%v", err)
		return nil, err
	}

	resource.WithPublished(r.Published)

	marshal, err := json.Marshal(&resource)
	if nil != err {
		logx.GetLogger("study").Errorf("PublishResource|Marshal Resource Error|%v", err)
		return nil, err
	}

	if err := ds.Set(ctx, infoKey, marshal, 0).Err(); err != nil {
		logx.GetLogger("study").Errorf("PublishResource|Set Resource Error|%v", err)
		return nil, err
	}

	info, err := r.QueryResourceInfo(ctx, ds)
	if err != nil {
		logx.GetLogger("study").Errorf("PublishResource|Query Resource Info Error|%v", err)
		return nil, err
	}

	logx.GetLogger("study").Infof("PublishResource|Update Resource Published Statuc Success|%v", common.ToStringWithoutError(info))
	return info, nil
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

	// 添加到文件的md5列表下面
	if err := ds.ZAdd(ctx, BuildMd5FileList(r.FileMd5), redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: r.Fid,
	}).Err(); err != nil {
		logx.GetLogger("study").Errorf("CreateUploadResource|Add Md5 File List Error|%v", err)
		return err
	}

	// 添加到章节的resource列表
	if err := ds.ZAdd(ctx, BuildChapterResourceList(r.Chid), redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: r.Fid,
	}).Err(); err != nil {
		logx.GetLogger("study").Errorf("CreateUploadResource|Add Chapter Resource List Error|%v", err)
		return err
	}

	return nil
}

func (r *Resource) DeleteResource(ctx context.Context, ds *redis.Client) error {
	key := BuildResourceInfo(r.Fid)
	// 删除md5下面的fid
	if err := ds.ZRem(ctx, BuildMd5FileList(r.FileMd5), r.Fid).Err(); err != nil {
		return err
	}
	if err := ds.Del(ctx, key).Err(); err != nil {
		return err
	}
	return nil
}

func (r *Resource) DeleteFormChapterList(ctx context.Context, ds *redis.Client) error {
	if err := ds.ZRem(ctx, BuildChapterResourceList(r.Chid), r.Fid).Err(); err != nil {
		return err
	}
	return nil
}

func (r *Resource) QueryResourceInfo(ctx context.Context, ds *redis.Client) (*Resource, error) {
	info := BuildResourceInfo(r.Fid)
	result, err := ds.Get(ctx, info).Result()
	if nil != err {
		return nil, err
	}

	err = json.Unmarshal([]byte(result), r)
	if nil != err {
		return nil, err
	}
	return r, nil
}

func (r *Resource) UpdateDownloadableResource(ctx context.Context, ds *redis.Client) (*Resource, error) {
	infoKey := BuildResourceInfo(r.Fid)
	result, err := ds.Get(ctx, infoKey).Result()
	if nil != err {
		logx.GetLogger("study").Errorf("UpdateDownloadableResource|Get Resource Error|%v", err)
		return nil, err
	}

	var resource Resource
	if err := json.Unmarshal([]byte(result), &resource); err != nil {
		logx.GetLogger("study").Errorf("UpdateDownloadableResource|Unmarshal Resource Info Error|%v", err)
		return nil, err
	}

	resource.WithDownloadable(r.Downloadable)

	raw, err := json.Marshal(&resource)
	if err != nil {
		logx.GetLogger("study").Errorf("UpdateDownloadableResource|Marshal Resource Error|%v", err)
		return nil, err
	}

	if err := ds.Set(ctx, infoKey, raw, 0).Err(); err != nil {
		logx.GetLogger("study").Errorf("UpdateDownloadableResource|Set Resource Error|%v", err)
		return nil, err
	}

	return &resource, nil
}

type ResourceList struct {
	ResourceList []Resource `json:"resource_list"`
	ReferFid     string     `json:"refer_fid"`
	Limit        int64      `json:"limit"`
	SourceId     string     `json:"source_id"`
}

func (rl *ResourceList) QueryResourceList(ctx context.Context, ds *redis.Client) (*ResourceList, error) {
	zrangeBy := redis.ZRangeBy{
		Min:    "0",
		Max:    strconv.Itoa(math.MaxInt64),
		Count:  rl.Limit,
		Offset: 0,
	}

	if len(rl.ReferFid) > 0 {
		result, err := ds.ZScore(ctx, BuildChapterResourceList(rl.SourceId), rl.ReferFid).Result()
		if nil != err {
			return nil, err
		}
		zrangeBy.Min = "(" + strconv.FormatInt(int64(result), 10)
	}

	result, err := ds.ZRangeByScore(ctx, BuildChapterResourceList(rl.SourceId), &zrangeBy).Result()
	if nil != err {
		return nil, err
	}

	for _, fid := range result {
		var resource Resource
		s, err := ds.Get(ctx, BuildResourceInfo(fid)).Result()
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(s), &resource); err != nil {
			return nil, err
		}
		rl.ResourceList = append(rl.ResourceList, resource)
	}
	return rl, nil
}
