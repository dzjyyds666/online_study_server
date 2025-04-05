package rpc

import (
	"context"
	"cos/api/config"
	pb "cos/api/proto"
	"cos/api/rpc/service"
	"errors"
	"fmt"
	"github.com/dzjyyds666/opensource/logx"
	"google.golang.org/grpc"
	"net"
)

func StartCosRpcServer(ctx context.Context) error {
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *config.GloableConfig.Host, *config.GloableConfig.RpcPort))
	if err != nil {
		logx.GetLogger("study").Errorf("StartCosRpcServer|Listen Error|%v", err)
		return err
	}

	cosServer := grpc.NewServer()
	pb.RegisterCosServer(cosServer, &service.CosService{})
	logx.GetLogger("study").Infof("gRPC Server is running on port %s", *config.GloableConfig.RpcPort)
	if err := cosServer.Serve(listen); err != nil {
		logx.GetLogger("study").Errorf("StartCosRpcServer|Serve Error|%v", err)
		return err
	}
	return errors.New("rpc service stop")
}
