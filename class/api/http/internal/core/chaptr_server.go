package core

import (
	"context"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/redis/go-redis/v9"
	"time"
)

type ChapterServer struct {
	ctx       context.Context
	chapterDB *redis.Client
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

	// 把章节信息保存到课程章节下面
	err = cs.chapterDB.ZAdd(ctx, BuildClassChapterList(*info.SourceId), redis.Z{
		Member: info.Chid,
		Score:  float64(time.Now().Unix()),
	}).Err()

	if err != nil {
		logx.GetLogger("study").Errorf("ChapterServer|CreateChapter|AddChapterToClassError|%v", err)
		return err
	}
	return nil
}
