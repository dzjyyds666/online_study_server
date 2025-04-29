package userRpcService

import (
	"bytes"
	"common/proto"
	"context"
	"io"
	"user/api/core"

	"github.com/dzjyyds666/opensource/common"
	"github.com/dzjyyds666/opensource/logx"
)

type UserRpcServer struct {
	UserServer *core.UserServer
	proto.UnimplementedUserServer
}

func (us *UserRpcServer) AddStudentToClass(ctx context.Context, req *proto.AddStudentToClassRequest) (*proto.UserCommonResponse, error) {
	logx.GetLogger("study").Infof("AddStudentToClass|succes|%s", common.ToStringWithoutError(req))
	err := us.UserServer.AddStudentToClass(context.Background(), req.Cid, req.Uid, req.Name)
	if err != nil {
		logx.GetLogger("study").Errorf("AddStudentToClass|AddStudentToClass|err:%v", err)
		return nil, err
	}
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

func (us *UserRpcServer) GetStudentsInfo(ctx context.Context, in *proto.StudentIds) (*proto.StudentInfos, error) {
	list := &proto.StudentInfos{Infos: make([]*proto.StudentInfo, 0, len(in.Uids))}
	for _, uid := range in.Uids {
		info, err := us.UserServer.QueryUserInfo(ctx, uid)
		if err != nil {
			logx.GetLogger("study").Errorf("GetStudentsInfo|QueryUserInfo|err:%v", err)
			continue
		}
		student := proto.StudentInfo{}
		student.Uid = info.Uid
		student.Name = info.Name
		student.College = info.Collage
		student.Major = info.Major
		student.Avatar = info.Avatar
		list.Infos = append(list.Infos, &student)
	}
	return list, nil
}

func (us *UserRpcServer) GetStudentClassList(ctx context.Context, in *proto.StudentIds) (*proto.ClassCids, error) {
	cids, err := us.UserServer.QueryStudentClassList(ctx, in.Uids[0])
	if err != nil {
		logx.GetLogger("study").Errorf("GetStudentClassList|QueryStudentClassList|err:%v", err)
		return nil, err
	}

	return &proto.ClassCids{
		Cids: cids,
	}, nil
}

func (us *UserRpcServer) GetUserInfo(ctx context.Context, in *proto.Uid) (*proto.UserInfo, error) {
	info, err := us.UserServer.QueryUserInfo(ctx, in.Uid)
	if err != nil {
		logx.GetLogger("study").Errorf("GetUserInfo|QueryUserInfo|err:%v", err)
		return nil, err
	}

	return &proto.UserInfo{
		Uid:      info.Uid,
		Username: info.Name,
		Avatar:   info.Avatar,
	}, nil
}
