package grpc

import (
	"context"
	"github.com/redis/go-redis/v9"
)

func StratClassGprcService(ctx context.Context, ds *redis.Client) error {
	//
	//listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *config.GloableConfig.Host, *config.GloableConfig.RpcPort))
	//if err!=nil{
	//	logx.GetLogger("study").Errorf("StratClassGprcService|Listen Error|%v", err)
	//	return err
	//}
	//
	//classGrpcServer := grpc.NewServer()
	//pb.RegisterCosServer(classGrpcServer, &server.ClassServer{Ds:ds})
	//return errors.New("class grpc service stop")

	return nil
}
