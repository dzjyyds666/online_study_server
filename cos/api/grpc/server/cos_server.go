package server

import (
	"context"
	pb "cos/api/grpc/proto"
	"github.com/dzjyyds666/opensource/common"
	"github.com/dzjyyds666/opensource/logx"
)

type CosServer struct {
	pb.UnimplementedCosServer
}

func (cs *CosServer) DeleteObject(ctx context.Context, req *pb.DeleteObjectRequest) (*pb.DeleteObjectResponse, error) {
	logx.GetLogger("study").Infof("receive delete object request: %v", common.ToStringWithoutError(req))

	return &pb.DeleteObjectResponse{
		Success: true,
	}, nil
}
