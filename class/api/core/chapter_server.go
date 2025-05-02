package core

import (
	"class/api/middleware"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"

	"github.com/dzjyyds666/opensource/logx"
	"github.com/redis/go-redis/v9"
)

type ChapterServer struct {
	ctx          context.Context
	chapterDB    *redis.Client
	mgCLi        *mongo.Collection
	resourceServ *ResourceServer
}

func NewChapterServer(ctx context.Context, dsClient *redis.Client, mongoCli *mongo.Client) *ChapterServer {

	server := NewResourceServer(ctx, dsClient, mongoCli)
	return &ChapterServer{
		ctx:          ctx,
		mgCLi:        mongoCli.Database("learnX").Collection("chapter"),
		chapterDB:    dsClient,
		resourceServ: server,
	}
}

func (cs *ChapterServer) QueryResourceList(ctx context.Context, list *ResourceList) error {
	return cs.resourceServ.QueryResourceList(ctx, list)
}

func (cs *ChapterServer) CreateChapter(ctx context.Context, info *Chapter) error {
	one, err := cs.mgCLi.InsertOne(ctx, info)
	if err != nil {
		lg.Errorf("ChapterServer|CreateChapter|CreateChapterError|%v", err)
		return err
	}
	if one.InsertedID == nil {
		lg.Errorf("CreateChapter|Insert Chapter Error|%v", err)
		return err
	}
	return nil
}

func (cs *ChapterServer) UpdateChapter(ctx context.Context, info *Chapter) error {
	id, err := cs.mgCLi.UpdateByID(ctx, info.Chid, bson.M{
		"$set": info,
	})
	if err != nil {
		lg.Errorf("UpdateChapter|Update Chapter Error|%v", err)
		return err
	}
	if id.ModifiedCount == 0 {
		lg.Errorf("UpdateChapter|Update Chapter Error|%v", err)
		return err
	}
	return nil
}

func (cs *ChapterServer) DeleteChapter(ctx context.Context, chid string) error {
	// 从课程的章节列表中删除
	err := cs.chapterDB.ZRem(ctx, BuildClassChapterList(chid), chid).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("ChapterServer|DeleteChapter|DeleteChapterFromClassError|%v", err)
		return err
	}
	// 先删除章节的信息
	one, err := cs.mgCLi.DeleteOne(ctx, bson.M{
		"_id": chid,
	})
	if err != nil {
		lg.Errorf("DeleteChapter|Delete Chapter Error|%v", err)
		return err
	}
	if one.DeletedCount == 0 {
		lg.Errorf("DeleteChapter|No Such Chapter")
		return nil
	}
	// 删除章节下面所有的资源信息
	resourceListKey := BuildChapterResourceList(chid)
	ids, err := cs.chapterDB.ZRange(ctx, resourceListKey, 0, -1).Result()
	if err != nil {
		logx.GetLogger("study").Errorf("ChapterServer|DeleteChapter|GetResourceListError|%v", err)
		return err
	}

	for _, id := range ids {
		// 删除资源信息
		_, err := cs.DeleteResource(ctx, id)
		if err != nil {
			logx.GetLogger("study").Errorf("ChapterServer|DeleteResource|DeleteResourceError|%v", err)
			break
		}
	}
	return nil
}

func (cs *ChapterServer) DeleteResource(ctx context.Context, fid string) (*Resource, error) {
	resourceInfo, err := cs.resourceServ.QueryResourceInfo(ctx, fid)
	if err != nil {
		logx.GetLogger("study").Errorf("ChapterServer|DeleteResource|QueryResourceInfoError|%v", err)
		return nil, err
	}

	// 先删除章节资源列表下的该资源
	err = cs.chapterDB.ZRem(ctx, BuildChapterResourceList(*resourceInfo.Chid), fid).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("ChapterServer|DeleteResource|DeleteResourceFromChapterError|%v", err)
		return nil, err
	}
	// 删除资源信息
	err = cs.resourceServ.DeleteResource(ctx, fid)
	if err != nil {
		logx.GetLogger("study").Errorf("ChapterServer|DeleteResource|DeleteResourceError|%v", err)
		return nil, err
	}
	return resourceInfo, nil
}

func (cs *ChapterServer) QueryChapterInfo(ctx context.Context, chid string, role int) (*Chapter, error) {
	var info Chapter
	err := cs.mgCLi.FindOne(ctx, bson.M{
		"_id": chid,
	}).Decode(&info)
	if err != nil {
		lg.Errorf("QueryChapterInfo|QueryChapterInfoError|%v", err)
		return nil, err
	}

	// 查询对应章节资源信息
	fids, err := cs.chapterDB.ZRange(ctx, BuildChapterResourceList(chid), 0, -1).Result()
	if err != nil {
		logx.GetLogger("study").Errorf("ChapterServer|QueryChapterInfo|QueryResourceListError|%v", err)
		return nil, err
	}

	for _, fid := range fids {
		resourceInfo, err := cs.resourceServ.QueryResourceInfo(ctx, fid)
		if err != nil {
			logx.GetLogger("study").Errorf("ChapterServer|QueryChapterInfo|QueryResourceInfoError|%v", err)
			return nil, err
		}
		if role == middleware.UserRole.Student {
			if !resourceInfo.IsPublished() {
				continue
			}
		}
		info.ResourceList = append(info.ResourceList, *resourceInfo)
	}
	return &info, nil
}

func (cs *ChapterServer) CreateResource(ctx context.Context, info *Resource) error {
	err := cs.resourceServ.CreateResource(ctx, info)
	if err != nil {
		logx.GetLogger("study").Errorf("ChapterServer|CreateResource|CreateResourceError|%v", err)
		return err
	}

	// 把资源添加到章节列表下面
	err = cs.chapterDB.ZAdd(ctx, BuildChapterResourceList(*info.Chid), redis.Z{
		Member: info.Fid,
		Score:  float64(time.Now().Unix()),
	}).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("ResourceServer|CreateResource|Add Chapter Resource List Error|%v", err)
		return err
	}
	return nil
}

func (cs *ChapterServer) UpdateResource(ctx context.Context, info *Resource) error {
	return cs.resourceServ.UpdateResource(ctx, info)
}
