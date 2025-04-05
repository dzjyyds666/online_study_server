package core

import (
	"context"
	"encoding/json"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/redis/go-redis/v9"
	"time"
)

type ChapterServer struct {
	ctx          context.Context
	chapterDB    *redis.Client
	resourceServ *ResourceServer
}

func NewChapterServer(ctx context.Context, dsClient *redis.Client) *ChapterServer {
	return &ChapterServer{
		ctx:       ctx,
		chapterDB: dsClient,
	}
}

func (cs *ChapterServer) CreateChapter(ctx context.Context, info *Chapter) error {
	rawData := info.Marshal()
	err := cs.chapterDB.Set(ctx, BuildChapterInfo(*info.Chid), rawData, 0).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("ChapterServer|CreateChapter|CreateChapterError|%v", err)
		return err
	}
	return nil
}

func (cs *ChapterServer) UpdateChapter(ctx context.Context, info *Chapter) error {
	// 查询章节的信息
	result, err := cs.chapterDB.Get(ctx, BuildChapterInfo(*info.Chid)).Result()
	if err != nil {
		logx.GetLogger("study").Errorf("ChapterServer|UpdateChapter|GetChapterInfoError|%v", err)
		return err
	}
	var oldInfo *Chapter
	err = json.Unmarshal([]byte(result), oldInfo)
	if err != nil {
		logx.GetLogger("study").Errorf("ChapterServer|UpdateChapter|UnmarshalChapterInfoError|%v", err)
		return err
	}

	oldInfo.ChapterName = info.ChapterName
	err = cs.chapterDB.Set(ctx, BuildChapterInfo(*info.Chid), oldInfo.Marshal(), 0).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("ChapterServer|UpdateChapter|UpdateChapterInfoError|%v", err)
		return err
	}
	return nil
}

func (cs *ChapterServer) DeleteChapter(ctx context.Context, chid string) error {
	// 先删除章节的信息
	chapterKey := BuildChapterInfo(chid)

	err := cs.chapterDB.Del(ctx, chapterKey).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("ChapterServer|DeleteChapter|DeleteChapterInfoError|%v", err)
		return err
	}
	// 删除章节下面所有的资源信息
	resourceListKey := BuildChapterResourceList(chid)
	ids, err := cs.chapterDB.ZRange(ctx, resourceListKey, 0, -1).Result()
	if err != nil {
		logx.GetLogger("study").Errorf("ChapterServer|DeleteChapter|GetResourceListError|%v", err)
		return err
	}

	for _, id := range ids {
		// 删除章节信息
		err := cs.DeleteResource(ctx, id, chid)
		if err != nil {
			logx.GetLogger("study").Errorf("ChapterServer|DeleteResource|DeleteResourceError|%v", err)
			break
		}
	}
	return nil
}

func (cs *ChapterServer) DeleteResource(ctx context.Context, fid string, chid string) error {
	// 先删除章节资源列表下的该资源
	err := cs.chapterDB.ZRem(ctx, BuildChapterResourceList(chid), fid).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("ChapterServer|DeleteResource|DeleteResourceFromChapterError|%v", err)
		return err
	}
	// 删除资源信息
	err = cs.resourceServ.DeleteResource(ctx, fid)
	if err != nil {
		logx.GetLogger("study").Errorf("ChapterServer|DeleteResource|DeleteResourceError|%v", err)
		return err
	}
	return nil
}

func (cs *ChapterServer) QueryChapterInfo(ctx context.Context, chid string) (*Chapter, error) {
	key := BuildChapterInfo(chid)
	result, err := cs.chapterDB.Get(ctx, key).Result()
	if err != nil {
		logx.GetLogger("study").Errorf("ChapterServer|QueryChapterInfo|QueryChapterInfoError|%v", err)
		return nil, err
	}

	var info *Chapter
	err = json.Unmarshal([]byte(result), info)
	if err != nil {
		logx.GetLogger("study").Errorf("ChapterServer|QueryChapterInfo|UnmarshalChapterInfoError|%v", err)
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

		info.ResourceList = append(info.ResourceList, *resourceInfo)
	}

	return info, nil
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
