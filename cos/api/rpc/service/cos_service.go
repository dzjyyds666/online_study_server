package service

import (
	"context"
	"cos/api/core"
	"cos/api/proto"
	"github.com/dzjyyds666/opensource/common"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/redis/go-redis/v9"
)

type CosService struct {
	cosDB     *redis.Client
	cosServer *core.CosFileServer
	proto.UnimplementedCosServer
}

func (cs *CosService) DeleteObject(ctx context.Context, req *proto.DeleteObjectRequest) (*proto.CommonResponse, error) {
	logx.GetLogger("study").Infof("receive delete object request: %v", common.ToStringWithoutError(req))

	return &proto.CommonResponse{
		Success: true,
	}, nil
}

func (cs *CosService) CopyObject(ctx context.Context, req *proto.CopyObjectRequest) (*proto.CommonResponse, error) {
	// 先查询出源文件的信息
	file, err := cs.cosServer.QueryCosFile(ctx, req.SrcFid)
	if err != nil {
		logx.GetLogger("study").Errorf("CosRpcService|CosService|CopyObject|QueryCosFileError|%v", err)
		return &proto.CommonResponse{Success: false}, nil
	}
	file.WithFid(req.DstFid)
	err = cs.cosServer.SaveFileInfo(ctx, file)
	if err != nil {
		logx.GetLogger("study").Errorf("CosRpcService|CopyObject|SaveFileInfoError|%v", err)
		return &proto.CommonResponse{Success: false}, nil
	}
	return &proto.CommonResponse{
		Success: true,
	}, nil
}
