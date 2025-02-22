package logic

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"github/dzjyyds666/online_study_server/user/api/internal/svc"
	"github/dzjyyds666/online_study_server/user/api/internal/types"
)

type HandlerLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHandlerLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandlerLoginLogic {
	return &HandlerLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HandlerLoginLogic) HandlerLogin(req *types.LoginReq) (resp *types.HttpResponse, err error) {
	name := "dzj"
	password := "123456"
	if req.Email == name && req.Password == password {
		resp = &types.HttpResponse{
			Code: 200,
			Data: types.LoginResp{
				Token:  "123456",
				UserId: "123456",
			},
			Msg: "success",
		}
	} else {
		resp = &types.HttpResponse{
			Code: 400,
			Data: "fail",
			Msg:  "fail",
		}
	}
	return
}
