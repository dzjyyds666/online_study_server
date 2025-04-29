package core

import (
	"common/proto"
	"common/rpc/client"
	"context"
	"github.com/dzjyyds666/opensource/common"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
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
	err = cs.rsDb.ZAdd(ctx, buildArticleCommentListKey(comment.ArticleId), redis.Z{
		Score:  float64(comment.CreateTs),
		Member: comment.Id,
	}).Err()
	if err != nil {
		lg.Errorf("CreateComment|ZAdd Error|%v", err)
		return err
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

func (cs *CommentServer) GetCommentList(ctx context.Context, list *ListComment) error {
	var commentIds []string
	if len(list.Uid) > 0 {
		// 获取用户的评论列表
		result, err := cs.rsDb.ZRange(ctx, buildUserCommentListKey(list.Uid), 0, -1).Result()
		if err != nil {
			lg.Errorf("GetCommentList|GetUserCommentListError|%v", err)
			return err
		}
		commentIds = result
	} else if len(list.ArticleId) > 0 {
		// 获取文章的评论列表
		result, err := cs.rsDb.ZRange(ctx, buildArticleCommentListKey(list.ArticleId), 0, -1).Result()
		if err != nil {
			lg.Errorf("GetCommentList|GetArticleCommentListError|%v", err)
			return err
		}
		commentIds = result
	}

	list.List = make([]*Comment, 0, len(commentIds))
	for _, commentId := range commentIds {
		var comment Comment
		// 从mongo中查询评论
		err := cs.mgDb.FindOne(ctx, bson.M{
			"_id": commentId,
		}).Decode(&comment)
		if err != nil {
			lg.Errorf("GetCommentList|FindOne Error|%v|%s", err, commentId)
			return err
		}

		user := client.GetUserRpcClient(ctx)
		info, err := user.GetUserInfo(ctx, &proto.Uid{Uid: comment.Author})
		if err != nil {
			lg.Errorf("GetCommentList|GetUserInfo Error|%v|%s", err, commentId)
			return err
		}

		comment.AuthorName = info.Username
		list.List = append(list.List, &comment)
	}

	lg.Infof("GetCommentList|GetCommentList|%v", common.ToStringWithoutError(list))

	return nil
}

func (cs *CommentServer) DeleteComment(ctx context.Context, cmid string) error {
	articleKey := buildArticleCommentListKey(cmid)
	err := cs.rsDb.ZRem(ctx, articleKey, cmid).Err()
	if err != nil {
		lg.Errorf("DeleteComment|ZRem Error|%v|%s", err, cmid)
		return err
	}
	userKey := buildUserCommentListKey(cmid)
	err = cs.rsDb.ZRem(ctx, userKey, cmid).Err()
	if err != nil {
		lg.Errorf("DeleteComment|ZRem Error|%v|%s", err, cmid)
		return err
	}
	// mongo删除评论
	_, err = cs.mgDb.DeleteOne(ctx, bson.M{
		"_id": cmid,
	})
	if err != nil {
		lg.Errorf("DeleteComment|DeleteOne Error|%v|%s", err, cmid)
	}
	return nil
}
