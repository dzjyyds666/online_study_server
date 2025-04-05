package client

import (
	"common/proto"
	"context"
	"google.golang.org/grpc"
)

var (
	userRpcClient     proto.UserClient
	userRpcClientConn *grpc.ClientConn
)

func GetUserRpcClient(ctx context.Context) proto.UserClient {
	if userRpcClientConn == nil {
		conn, err := grpc.DialContext(ctx, "127.0.0.1:8080", grpc.WithInsecure())
		if err != nil {
			panic(err)
		}
		userRpcClientConn = conn
		userRpcClient = proto.NewUserClient(conn)
	}
	return userRpcClient
}

func CloseUserRpcClient() {
	if userRpcClientConn != nil {
		userRpcClientConn.Close()
	}
}
