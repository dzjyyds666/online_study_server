package userRpcService

import (
	"bytes"
	"common/proto"
	"context"
	"github.com/dzjyyds666/opensource/logx"
	"io"
	"user/api/core"
)

type UserRpcServer struct {
	UserServer *core.UserServer
	proto.UnimplementedUserServer
}

func (us *UserRpcServer) AddStudentToClass(ctx context.Context, req *proto.AddStudentToClassRequest) (*proto.UserCommonResponse, error) {
	return &proto.UserCommonResponse{Success: true}, nil
}

// 批量注册用户
func (us *UserRpcServer) BatchAddStudentToClass(stream proto.User_BatchAddStudentToClassServer) error {
	var fileData bytes.Buffer
	var fileName string
	var cid string
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			logx.GetLogger("study").Errorf("BatchRegister|Recv Error|%v", err)
			return err
		}
		if len(fileName) <= 0 {
			fileName = chunk.Filename
		}
		if len(cid) <= 0 {
			cid = chunk.Cid
		}
		fileData.Write(chunk.Content)
	}
	ids, err := us.UserServer.BetchAddStudentToClass(context.Background(), cid, &fileData)
	if err != nil {
		logx.GetLogger("study").Errorf("BatchRegister|BetchAddStudentToClass Error|%v", err)
		return err
	}
	return stream.SendAndClose(&proto.StudentIds{
		Uids: ids,
	})
}
