package rpc

import (
	"common/proto"
	"context"
	"errors"
	"fmt"
	"github.com/dzjyyds666/opensource/logx"
	"google.golang.org/grpc"
	"net"
	"user/api/config"
	"user/api/core"
	"user/api/rpc/service"
)

func StratUserRpcServer(ctx context.Context, server *core.UserServer) error {
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *config.GloableConfig.Host, *config.GloableConfig.RpcPort))
	if err != nil {
		logx.GetLogger("study").Errorf("StratUserRpcServer|Listen Error|%v", err)
		return err
	}

	userServer := grpc.NewServer()
	proto.RegisterUserServer(userServer, &userRpcService.UserRpcServer{
		UserServer: server,
	})
	if err := userServer.Serve(listen); err != nil {
		logx.GetLogger("study").Errorf("StratUserRpcServer|Serve Error|%v", err)
		return err
	}
	return errors.New("rpc service stop")
}
