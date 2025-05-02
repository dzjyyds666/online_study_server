package core

import (
	"common/proto"
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"math"
	"strconv"
)

const (
	RedisResourceKey = "class:resource:%s" // fid
)

func BuildResourceKey(fid string) string {
	return fmt.Sprintf(RedisResourceKey, fid)
}

type Resource struct {
	Fid          *string             `json:"fid,omitempty" bson:"_id"`
	Published    *bool               `json:"published" bson:"published"`
	Downloadable *bool               `json:"downloadable" bson:"downloadable"` // 是否可以下载
	Chid         *string             `json:"chid,omitempty" bson:"chid"`
	FileInfo     *proto.ResourceInfo `json:"file_info,omitempty" bson:"fileInfo,omitempty"` // 文件信息
}

func (r *Resource) WithFileInfo(info *proto.ResourceInfo) *Resource {
	r.FileInfo = info
	return r
}

func (r *Resource) WithFid(fid string) *Resource {
	r.Fid = &fid
	return r
}

func (r *Resource) WithChid(chid string) *Resource {
	r.Chid = &chid
	return r
}

func (r *Resource) WithPublished(published bool) *Resource {
	r.Published = &published
	return r
}

func (r *Resource) WithDownloadable(downloadable bool) *Resource {
	r.Downloadable = &downloadable
	return r
}

func (r *Resource) Marshal() string {
	raw, _ := json.Marshal(r)
	return string(raw)
}

func (r *Resource) IsPublished() bool {
	return *r.Published == true
}

func (r *Resource) IsDownloadable() bool {
	return *r.Downloadable == true
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
