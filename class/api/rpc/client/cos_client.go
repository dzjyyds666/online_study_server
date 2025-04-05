package client

import (
	"common/proto"
	"context"
	"github.com/dzjyyds666/opensource/logx"
	"google.golang.org/grpc"
	"sync"
)

var (
	cosRpcClient  *proto.CosClient
	once          sync.Once
	cosClientConn *grpc.ClientConn
)

func GetCosRpcClient(ctx context.Context) *proto.CosClient {
	once.Do(func() {
		var err error
		// todo 修改rpc启动从配置文件中读取
		cosClientConn, err = grpc.DialContext(ctx, "127.0.0.1:29002", grpc.WithInsecure())
		if err != nil {
			logx.GetLogger("study").Errorf("UserServer|StartError|NewUserServer|err:%v", err)
			panic(err)
		}
		client := proto.NewCosClient(cosClientConn)
		cosRpcClient = &client
	})
	return cosRpcClient
}

func CloseCosRpcClient() {
	if cosClientConn != nil {
		cosClientConn.Close()
	}
}
