package userRpcService

import (
	"context"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"user/api/proto"
)

type UserServer struct {
	Ds *redis.Client
	Ms *gorm.DB
	proto.UnimplementedUserServer
}

func (us *UserServer) AddStudentToClass(ctx context.Context, req *proto.AddStudentToClassRequest) (*proto.AddStudentToClassResponse, error) {
	return &proto.AddStudentToClassResponse{Success: true}, nil
}
