package core

import (
	"context"
	"github.com/redis/go-redis/v9"
)

type ClassResource struct {
	Fid      string `json:"fid"`
	FileName string `json:"file_name"`
	FileType string `json:"file_type"`
	FileSize int64  `json:"file_size"`
	UploadTs int64  `json:"upload_ts"`
}

func (r *ClassResource) WithFid(id string) *ClassResource {
	r.Fid = id
	return r
}

func (r *ClassResource) WithFileName(name string) *ClassResource {
	r.FileName = name
	return r
}

func (r *ClassResource) WithFileType(type_ string) *ClassResource {
	r.FileType = type_
	return r
}

func (r *ClassResource) WithFileSize(size int64) *ClassResource {
	r.FileSize = size
	return r
}

func (r *ClassResource) WithUploadTs(ts int64) *ClassResource {
	r.UploadTs = ts
	return r
}

func (r *ClassResource) CreateResource(ctx context.Context, ds *redis.Client) error {

}
