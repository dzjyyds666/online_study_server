package core

import (
	"context"
	"encoding/json"
	"time"

	"github.com/dzjyyds666/opensource/logx"
	"github.com/redis/go-redis/v9"
)

type ChapterServer struct {
	ctx          context.Context
	chapterDB    *redis.Client
	resourceServ *ResourceServer
}

func NewChapterServer(ctx context.Context, dsClient *redis.Client) *ChapterServer {

	server := NewResourceServer(ctx, dsClient)
	return &ChapterServer{
		ctx:          ctx,
		chapterDB:    dsClient,
		resourceServ: server,
	}
}

func (cs *ChapterServer) QueryResourceList(ctx context.Context, list *ResourceList) error {
	return cs.resourceServ.QueryResourceList(ctx, list)
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
	// 从课程的章节列表中删除
	err := cs.chapterDB.ZRem(ctx, BuildClassChapterList(chid), chid).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("ChapterServer|DeleteChapter|DeleteChapterFromClassError|%v", err)
		return err
	}
	// 先删除章节的信息
	chapterKey := BuildChapterInfo(chid)

	err = cs.chapterDB.Del(ctx, chapterKey).Err()
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

func (cs *ChapterServer) QueryChapterInfo(ctx context.Context, chid string) (*Chapter, error) {
	key := BuildChapterInfo(chid)
	result, err := cs.chapterDB.Get(ctx, key).Result()
	if err != nil {
		logx.GetLogger("study").Errorf("ChapterServer|QueryChapterInfo|QueryChapterInfoError|%v", err)
		return nil, err
	}

	var info Chapter
	err = json.Unmarshal([]byte(result), &info)
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

func (cs *ChapterServer) UpdateResource(ctx context.Context, info *Resource) (*Resource, error) {
	return cs.resourceServ.UpdateResource(ctx, info)
}
