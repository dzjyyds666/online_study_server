package core

import (
	"context"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

type PlateServer struct {
	ctx  context.Context
	rsDb *redis.Client
	mgDb *mongo.Client
}

func NewPlateServer(ctx context.Context, rsDb *redis.Client, mgDb *mongo.Client) *PlateServer {
	return &PlateServer{
		ctx:  ctx,
		rsDb: rsDb,
		mgDb: mgDb,
	}
}
