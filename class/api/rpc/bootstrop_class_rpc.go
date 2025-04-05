package rpc

import (
	"common/rpc/client"
	"context"
	"github.com/redis/go-redis/v9"
)

func StratClassRpcService(ctx context.Context, ds *redis.Client) error {
	//
	//listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *config.GloableConfig.Host, *config.GloableConfig.RpcPort))
	//if err!=nil{
	//	logx.GetLogger("study").Errorf("StratClassRpcService|Listen Error|%v", err)
	//	return err
	//}
	//
	//classGrpcServer := rpc.NewServer()
	//pb.RegisterCosServer(classGrpcServer, &service.ClassServer{Ds:ds})
	//return errors.New("class rpc service stop")

	return nil
}

func StopClassRpcService() {
	client.CloseCosRpcClient()
}
