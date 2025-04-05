package core

import (
	"context"
	"encoding/json"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/redis/go-redis/v9"
)

type CosFileServer struct {
	cosDB *redis.Client
	ctx   context.Context
}

func NewCosFileServer(ctx context.Context, cosDB *redis.Client) *CosFileServer {
	return &CosFileServer{
		cosDB: cosDB,
		ctx:   ctx,
	}
}

func (cfs *CosFileServer) QueryCosFile(ctx context.Context, fid string) (*CosFile, error) {
	key := buildFileInfoKey(fid)
	result, err := cfs.cosDB.Get(ctx, key).Result()
	if err != nil {
		logx.GetLogger("study").Errorf("CosFileServer|QueryCosFile|GetCosFileInfoError|%v", err)
		return nil, err
	}

	var cosFile CosFile
	err = json.Unmarshal([]byte(result), &cosFile)
	if err != nil {
		logx.GetLogger("study").Errorf("CosFileServer|QueryCosFile|UnmarshalError|%v", err)
		return nil, err
	}
	return &cosFile, nil
}

func (cfs *CosFileServer) SaveFileInfo(ctx context.Context, cosFile *CosFile) error {
	key := buildFileInfoKey(*cosFile.Fid)
	err := cfs.cosDB.Set(ctx, key, cosFile.Marshal(), 0).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("CosFileServer|CopyNewFile|SetError|%v", err)
		return err
	}
	return nil
}
