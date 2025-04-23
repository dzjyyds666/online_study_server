package communityService

import (
	"community/api/core"
	"context"
)

type CommunityService struct {
	ctx       context.Context
	plateServ *core.PlateServer
}

func NewCommunityService(ctx context.Context, plate *core.PlateServer) *CommunityService {
	return &CommunityService{
		ctx:       ctx,
		plateServ: plate,
	}
}
