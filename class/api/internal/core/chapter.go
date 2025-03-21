package core

import (
	"context"
	"encoding/json"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/redis/go-redis/v9"
	"time"
)

// 章节
type Chapter struct {
	Chid        *string `json:"chid"`
	ChapterName *string `json:"chapter_name"`
	CreateTs    *int64  `json:"create_ts"`
}

func (ch *Chapter) WithChid(id string) *Chapter {
	ch.Chid = &id
	return ch
}

func (ch *Chapter) WithChapterName(name string) *Chapter {
	ch.ChapterName = &name
	return ch
}

func (ch *Chapter) WithCreateTs(ts int64) *Chapter {
	ch.CreateTs = &ts
	return ch
}

func (ch *Chapter) RanameChapter(ctx context.Context, newName string, ds *redis.Client) error {
	chapterKey := BuildChapterInfo(*ch.Chid)
	result, err := ds.Get(ctx, chapterKey).Result()
	if err != nil {
		logx.GetLogger("study").Errorf("RanameChapter|Get Chapter Error|%v", err)
		return err
	}
	err = json.Unmarshal([]byte(result), ch)
	if err != nil {
		logx.GetLogger("study").Errorf("RanameChapter|Unmarshal Chapter Error|%v", err)
		return err
	}

	ch.ChapterName = &newName

	marshal, err := json.Marshal(ch)
	if err != nil {
		logx.GetLogger("study").Errorf("CreateChapter|Marshal Chapter Error|%v", err)
		return err
	}
	_, err = ds.Set(ctx, chapterKey, string(marshal), 0).Result()
	if err != nil {
		logx.GetLogger("study").Errorf("CreateChapter|Set Chapter Error|%v", err)
		return err
	}
	return nil
}

func (ci *Chapter) CreateChapter(ctx context.Context, cid string, ds *redis.Client) error {
	// 从reids中获取到class的indexinfo
	chapterStr, _ := json.Marshal(ci)

	classKey := BuildClassChapterList(cid)
	_, err := ds.ZAdd(ctx, classKey, redis.Z{
		Member: ci.Chid,
		Score:  float64(time.Now().Unix()),
	}).Result()

	if err != nil {
		logx.GetLogger("study").Errorf("CreateChapter|Add Chapter Error|%v", err)
		return err
	}

	// 把章节的信息存储到reids中
	chapterKey := BuildChapterInfo(*ci.Chid)
	_, err = ds.Set(ctx, chapterKey, string(chapterStr), 0).Result()
	if err != nil {
		logx.GetLogger("study").Errorf("CreateChapter|Set Chapter Info Error|%v", err)
		return err
	}
	return nil
}
