package main

import (
	"flag"
	"fmt"
	"github.com/dzjyyds666/opensource/logx"

	"github.com/dzjyyds666/online_study_server/object_storage/api/internal/config"
	"github.com/dzjyyds666/online_study_server/object_storage/api/internal/handler"
	"github.com/dzjyyds666/online_study_server/object_storage/api/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/oss-api.yaml", "the config file")
var bucketPolicyFile = flag.String("b", "etc/policy.json", "the bucket policy file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	_, err := c.S3.WithPolicy(*bucketPolicyFile)
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("load bucket policy error:%v", err)
		return
	}

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
