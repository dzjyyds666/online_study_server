package core

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"net/http"
	"time"
)

type LambdaQueueServer struct {
	ctx     context.Context
	hcli    *http.Client
	cosServ *CosFileServer
}

func NewLambdaQueueServer(ctx context.Context, hcli *http.Client) *LambdaQueueServer {
	return &LambdaQueueServer{
		ctx:     ctx,
		hcli:    hcli,
		cosServ: NewCosFileServer(ctx, nil, nil),
	}
}

func (lqs *LambdaQueueServer) WatchQueue(ctx context.Context, queueName string) error {
	// 开始监听队列
	for {
		result, err := lqs.cosServ.cosDB.BRPop(ctx, 5*time.Second, queueName).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				continue
			} else {
				panic(err)
			}
		}
		//获取到fid，开启线程进行处理，限制线程的数量
	}
}
