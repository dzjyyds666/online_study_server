package logic

import (
	"context"

	"github.com/dzjyyds666/online_study_server/object_storage/api/internal/svc"
	"github.com/dzjyyds666/online_study_server/object_storage/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ApplyUploadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewApplyUploadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApplyUploadLogic {
	return &ApplyUploadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ApplyUploadLogic) ApplyUpload(req *types.ApplyUploadRequest) (resp *types.HttpResponse, err error) {
	
	return
}
