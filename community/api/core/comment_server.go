package core

import (
	"context"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type CommentServer struct {
	ctx  context.Context
	rsDb *redis.Client
	mgDb *mongo.Collection
}

func NewCommentServer(ctx context.Context, rsDb *redis.Client, mgDb *mongo.Client) *CommentServer {
	return &CommentServer{
		ctx:  ctx,
		rsDb: rsDb,
		mgDb: mgDb.Database("learnX").Collection("comment"),
	}
}

func (cs *CommentServer) CreateComment(ctx context.Context, comment *Comment) error {
	comment.WithId(newCommentId(8)).
		WithCreateTs(time.Now().Unix()).
		WithDisLikes(0).
		WithLikes(0)
	_, err := cs.mgDb.InsertOne(ctx, comment)
	if err != nil {
		lg.Errorf("CreateComment|Insert Comment Error|%v", err)
		return err
	}
	// 判断是不是子评论，子评论不存入索引
	if len(comment.ParentId) <= 0 {
		err := cs.rsDb.ZAdd(ctx, buildArticleCommentListKey(comment.ArticleId), redis.Z{
			Score:  float64(comment.CreateTs),
			Member: comment.Id,
		}).Err()
		if err != nil {
			lg.Errorf("CreateComment|ZAdd Error|%v", err)
			return err
		}
	}
	// 存入用户自身的评论列表中
	err = cs.rsDb.ZAdd(ctx, buildUserCommentListKey(comment.Author), redis.Z{
		Score:  float64(comment.CreateTs),
		Member: comment.Id,
	}).Err()
	if err != nil {
		lg.Errorf("CreateComment|ZAdd Error|%v", err)
		return err
	}
	return nil
}
