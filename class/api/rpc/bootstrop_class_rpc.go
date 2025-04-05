package rpc

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
	//classGrpcServer := rpc.NewServer()
	//pb.RegisterCosServer(classGrpcServer, &service.ClassServer{Ds:ds})
	//return errors.New("class rpc service stop")

	return nil
}
