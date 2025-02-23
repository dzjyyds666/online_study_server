package svc

import (
	"github.com/dzjyyds666/online_study_server/object_storage/api/internal/config"
)

type ServiceContext struct {
	Config config.Config
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
	}
}
