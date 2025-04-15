package core

import (
	"common/proto"
	"common/rpc/client"
	"context"
	"time"

	"github.com/dzjyyds666/opensource/logx"
	"github.com/redis/go-redis/v9"
)

type ClassStudentServer struct {
	ctx       context.Context
	studentDB *redis.Client
}

func NewClassStudentServer(ctx context.Context, dsClient *redis.Client) *ClassStudentServer {
	return &ClassStudentServer{
		ctx:       ctx,
		studentDB: dsClient,
	}
}

func (css *ClassStudentServer) AddStudentToClass(ctx context.Context, uid string, cid string) error {
	// TODO rpc服务判断学生是否注册，如果注册过的话直接把学生添加到班级列表中，没有注册过的话，就先创建学生，再把学生添加到半截列表中
	list := BuildClassStudentList(cid)
	err := css.studentDB.ZAdd(ctx, list, redis.Z{
		Member: uid,
		Score:  float64(time.Now().Unix()),
	}).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("AddStudentToClass|Add Student To Class Error|%v", err)
		return err
	}
	return nil
}

func (css *ClassStudentServer) QueryStudentList(ctx context.Context, cid string) (*proto.StudentInfos, error) {
	key := BuildClassStudentList(cid)
	uids, err := css.studentDB.ZRange(ctx, key, 0, -1).Result()
	if err != nil {
		logx.GetLogger("study").Errorf("QueryStudentList|Query Student List Error|%v", err)
		return nil, err
	}
	userClient := client.GetUserRpcClient(ctx)
	resp, err := userClient.GetStudentsInfo(ctx, &proto.StudentIds{Uids: uids})
	if err != nil {
		logx.GetLogger("study").Errorf("QueryStudentList|GetStudentsInfo|%v", err)
		return nil, err
	}

	return resp, nil
}
