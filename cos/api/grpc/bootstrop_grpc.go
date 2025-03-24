package grpc

import (
	"context"
	"cos/api/config"
	pb "cos/api/grpc/proto"
	"cos/api/grpc/server"
	"errors"
	"fmt"
	"github.com/dzjyyds666/opensource/logx"
	"google.golang.org/grpc"
	"net"
)

func StartRpcService(ctx context.Context) error {
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *config.GloableConfig.Host, *config.GloableConfig.RpcPort))
	if err != nil {
		logx.GetLogger("study").Errorf("StartRpcService|Listen Error|%v", err)
		return err
	}

	cosServer := grpc.NewServer()
	pb.RegisterCosServer(cosServer, &server.CosServer{})
	logx.GetLogger("study").Infof("gRPC Server is running on port %s", *config.GloableConfig.RpcPort)
	if err := cosServer.Serve(listen); err != nil {
		logx.GetLogger("study").Errorf("StartRpcService|Serve Error|%v", err)
		return err
	}
	return errors.New("grpc service stop")
}
