package userRpcService

import (
	"context"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	pb "user/api/rpc/proto"
)

type UserServer struct {
	Ds *redis.Client
	Ms *gorm.DB
	pb.UnimplementedUserServer
}

func (us *UserServer) AddStudentToClass(ctx context.Context, req *pb.AddStudentToClassRequest) (*pb.AddStudentToClassResponse, error) {
	return &pb.AddStudentToClassResponse{Success: true}, nil
}
