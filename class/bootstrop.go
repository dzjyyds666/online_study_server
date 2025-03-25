package main

import (
	"class/api/grpc"
	"class/api/http"
	"github.com/dzjyyds666/opensource/logx"
	"golang.org/x/sync/errgroup"
)

func main() {
	var g errgroup.Group
	g.Go(func() error {
		err := http.StartClassHttpServer()
		if nil != err {
			logx.GetLogger("study").Errorf("main|StartApiServer|err:%v", err)
			return err
		}
		return nil
	})
	g.Go(func() error {
		err := grpc.StartRpcService()
		if nil != err {
			logx.GetLogger("study").Errorf("main|StartRpcServer|err:%v", err)
			return err
		}
		return nil
	})
	if err := g.Wait(); err != nil {
		logx.GetLogger("study").Errorf("main|err:%v", err)
	}
}
