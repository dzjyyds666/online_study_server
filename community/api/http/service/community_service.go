package communityService

import (
	"context"
	"github.com/redis/go-redis/v9"
)

type CommunityService struct {
	ctx         context.Context
	communityDb *redis.Client
}

func NewCommunityService(ctx context.Context, dsClient *redis.Client) *CommunityService {
	return &CommunityService{
		ctx:         ctx,
		communityDb: dsClient,
	}
}
