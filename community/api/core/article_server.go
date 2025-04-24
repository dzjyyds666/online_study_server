package core

import (
	"context"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type ArticleServer struct {
	ctx         context.Context
	rsDb        *redis.Client
	articleMgDb *mongo.Collection
}

func NewArticleServer(ctx context.Context, rsDb *redis.Client, mgDb *mongo.Client) *ArticleServer {
	return &ArticleServer{
		ctx:         ctx,
		rsDb:        rsDb,
		articleMgDb: mgDb.Database("learnX").Collection("article"),
	}
}

func (as *ArticleServer) CreateArticle(ctx context.Context, article *Article) error {
	article.WithId(newArticleId(8)).
		WithCreateTs(time.Now().Unix()).
		WithUpdateTs(time.Now().Unix()).
		WithPublished(false)
	_, err := as.articleMgDb.InsertOne(ctx, article)
	if err != nil {
		lg.Errorf("CreateArticle|Insert Article Error|%v", err)
		return err
	}

	// 存入到待审核列表中
	key := buildPlateArticleAuditListKey(article.PlateId)
	err = as.rsDb.ZAdd(ctx, key, redis.Z{
		Member: article.Id,
		Score:  float64(article.CreateTs),
	}).Err()
	if err != nil {
		lg.Errorf("CreateArticle|ZAdd Error|%v", err)
		return err
	}
	return nil
}
