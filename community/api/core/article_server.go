package core

import (
	"context"
	"github.com/dzjyyds666/opensource/common"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
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
		WithStatus(ArticleStatuses.Audit)
	_, err := as.articleMgDb.InsertOne(ctx, article)
	if err != nil {
		lg.Errorf("CreateArticle|Insert Article Error|%v", err)
		return err
	}

	// 存入用户自身的列表下面
	err = as.rsDb.ZAdd(ctx, buildUserArticleListKey(article.Author), redis.Z{
		Member: article.Id,
		Score:  float64(article.CreateTs),
	}).Err()
	if err != nil {
		lg.Errorf("CreateArticle|ZAdd Error|%v", err)
		return err
	}

	// 存入到待审核列表中
	key := buildArticleAuditListKey()
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

func (as *ArticleServer) UpdateArticle(ctx context.Context, article *Article) error {
	if len(article.Status) > 0 {
		// 更新状态，
		switch article.Status {
		case ArticleStatuses.Published:
			// 从审核队列中取消，移动到对应的plate列表下面
			key := buildArticleAuditListKey()
			err := as.rsDb.ZRem(ctx, key, article.Id).Err()
			if err != nil {
				lg.Errorf("UpdateArticle|ZRem Error|%v", err)
				return err
			}
			err = as.rsDb.ZAdd(ctx, buildPlateArticleListKey(article.PlateId), redis.Z{
				Member: article.Id,
				Score:  float64(article.CreateTs),
			}).Err()
			if err != nil {
				lg.Errorf("UpdateArticle|ZAdd Error|%v", err)
				return err
			}
		case ArticleStatuses.Illegal:
			// 从审核队列中取消
			key := buildArticleAuditListKey()
			err := as.rsDb.ZRem(ctx, key, article.Id).Err()
			if err != nil {
				lg.Errorf("UpdateArticle|ZRem Error|%v", err)
				return err
			}
		}
	}

	filter := bson.M{
		"_id": article.Id,
	}

	update := bson.M{
		"$set": article,
	}

	result, err := as.articleMgDb.UpdateOne(ctx, filter, update)
	if err != nil {
		lg.Errorf("UpdateArticle|UpdateOne Error|%v", err)
		return err
	}

	if result.MatchedCount == 0 {
		lg.Errorf("UpdateArticle|No Data Match|%v", common.ToStringWithoutError(result))
		return ErrNoMatchData
	}
	lg.Errorf("UpdateArticle|UpdateArticle Success|%v", common.ToStringWithoutError(result))
	return nil
}

func (as *ArticleServer) DeleteArticle(ctx context.Context, articleId string) error {
	// todo 删除文章，比较麻烦
	return nil
}

func (as *ArticleServer) ListArticle(ctx context.Context, list *ListArticle) error {
	articleIds := make([]string, 0)
	if len(list.PlateId) > 0 {
		// 查询板块下的文章
		result, err := as.rsDb.ZRange(ctx, buildPlateArticleListKey(list.PlateId), 0, -1).Result()
		if err != nil {
			lg.Errorf("ListArticle|ZRangeByScore Error|%v", err)
			return err
		}
		articleIds = result
	} else if len(list.Uid) > 0 {
		// 查询用户下的文章
		result, err := as.rsDb.ZRange(ctx, buildUserArticleListKey(list.Uid), 0, -1).Result()
		if err != nil {
			lg.Errorf("ListArticle|ZRangeByScore Error|%v", err)
			return err
		}
		articleIds = result
	} else if list.Audit == true {
		// 查询审核中的文章
		result, err := as.rsDb.ZRange(ctx, buildArticleAuditListKey(), 0, -1).Result()
		if err != nil {
			lg.Errorf("ListArticle|ZRangeByScore Error|%v", err)
			return err
		}
		articleIds = result
	}
	list.List = make([]*Article, 0, len(articleIds))
	for _, articleId := range articleIds {
		var article Article
		err := as.articleMgDb.FindOne(ctx, bson.M{
			"_id": articleId,
		}).Decode(&article)
		if err != nil {
			lg.Errorf("ListArticle|Find Error|%v", err)
			return err
		}
		list.List = append(list.List, &article)
	}
	return nil
}
