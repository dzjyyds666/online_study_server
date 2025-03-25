package rpc

import (
	"context"
	"errors"
	"fmt"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"gorm.io/gorm"
	"net"
	"user/api/config"
	"user/api/rpc/internal/server"
	pb "user/api/rpc/proto"
)

func StratUserRpcServer(ctx context.Context, redis *redis.Client, mysql *gorm.DB) error {
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *config.GloableConfig.Host, *config.GloableConfig.RpcPort))
	if err != nil {
		logx.GetLogger("study").Errorf("StratUserRpcServer|Listen Error|%v", err)
		return err
	}

	userServer := grpc.NewServer()

	pb.RegisterUserServer(userServer, &userRpcService.UserServer{
		Ds: redis,
		Ms: mysql,
	})

	if err := userServer.Serve(listen); err != nil {
		logx.GetLogger("study").Errorf("StratUserRpcServer|Serve Error|%v", err)
		return err
	}

	return errors.New("grpc service stop")
}
