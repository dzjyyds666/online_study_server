package core

import (
	"context"
	"encoding/json"
	"github.com/dzjyyds666/opensource/common"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/redis/go-redis/v9"
	"math"
	"strconv"
	"time"
)

// 章节
type Chapter struct {
	Chid         *string    `json:"chid,omitempty"`
	ChapterName  *string    `json:"chapter_name,omitempty"`
	SourceId     *string    `json:"source_id,omitempty"`
	ResourceList []Resource `json:"resource_list,omitempty"`
}

type ChapterList struct {
	SourceId    string    `json:"source_id"`
	ReferChid   string    `json:"refer_chid"`
	Limit       int64     `json:"limit"`
	ChapterList []Chapter `json:"chapter_list"`
}

func (ch *Chapter) WithChid(id string) *Chapter {
	ch.Chid = &id
	return ch
}

func (ch *Chapter) WithChapterName(name string) *Chapter {
	ch.ChapterName = &name
	return ch
}

func (ch *Chapter) WithSourceId(id string) *Chapter {
	ch.SourceId = &id
	return ch
}

func (ch *Chapter) RanameChapter(ctx context.Context, ds *redis.Client) (*Chapter, error) {
	chapterKey := BuildChapterInfo(*ch.Chid)
	result, err := ds.Get(ctx, chapterKey).Result()
	if err != nil {
		logx.GetLogger("study").Errorf("RanameChapter|Get Chapter Error|%v", err)
		return nil, err
	}

	var info Chapter
	err = json.Unmarshal([]byte(result), &info)
	if err != nil {
		logx.GetLogger("study").Errorf("RanameChapter|Unmarshal Chapter Error|%v", err)
		return nil, err
	}

	info.ChapterName = ch.ChapterName

	marshal, err := json.Marshal(info)
	if err != nil {
		logx.GetLogger("study").Errorf("CreateChapter|Marshal Chapter Error|%v", err)
		return nil, err
	}
	_, err = ds.Set(ctx, chapterKey, string(marshal), 0).Result()
	if err != nil {
		logx.GetLogger("study").Errorf("CreateChapter|Set Chapter Error|%v", err)
		return nil, err
	}
	logx.GetLogger("study").Infof("RanameChapter|RanameChapter Success|%s", common.ToStringWithoutError(info))
	return &info, err
}

func (ci *Chapter) CreateChapter(ctx context.Context, cid string, ds *redis.Client) error {
	// 从reids中获取到class的indexinfo
	chapterStr, _ := json.Marshal(ci)

	classKey := BuildSourceChapterList(cid)
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

func (ci *Chapter) DeleteChapter(ctx context.Context, ds *redis.Client) error {

	infoKey := BuildChapterInfo(*ci.Chid)
	result, err := ds.Get(ctx, infoKey).Result()
	if nil != err {
		logx.GetLogger("study").Errorf("DeleteChapter|Get Chapter Info Error|%v", err)
		return err
	}

	err = json.Unmarshal([]byte(result), ci)
	if err != nil {
		logx.GetLogger("study").Errorf("DeleteChapter|Unmarshal Chapter Info Error|%v", err)
		return err
	}

	// 从source的列表中删除
	sourceKey := BuildSourceChapterList(*ci.SourceId)
	err = ds.ZRem(ctx, sourceKey, ci.Chid).Err()
	if nil != err {
		logx.GetLogger("study").Errorf("DeleteChapter|Delete Source Chapter Error|%v", err)
		return err
	}

	// 遍历删除章节下面的所有的资源
	referFid := ""
	for {
		list, err := ci.QueryResourcList(ctx, ds, referFid, 100)
		if nil != err {
			logx.GetLogger("study").Errorf("DeleteChapter|Query Resourc List Error|%v", err)
			return err
		}

		for _, fid := range list {
			resource := Resource{Fid: fid}
			info, err := resource.QueryResourceInfo(ctx, ds)
			if nil != err {
				logx.GetLogger("study").Errorf("DeleteChapter|Query Resource Info Error|%v", err)
				return err
			}

			err = info.DeleteResource(ctx, ds)
			if nil != err {
				logx.GetLogger("study").Errorf("DeleteChapter|Delete Resource Error|%v", err)
				return err
			}
		}

		if len(list) < 100 {
			break
		} else if len(list) == 100 {
			referFid = list[len(list)-1]
		}
	}

	// 删除章节对应的资源列表
	if err = ds.Del(ctx, BuildChapterResourceList(*ci.Chid)).Err(); err != nil {
		logx.GetLogger("study").Errorf("DeleteChapter|Delete Chapter Resource List Error|%v", err)
		return err
	}

	// 删除chapter的信息
	err = ds.Del(ctx, infoKey).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("DeleteChapter|Delete Chapter Info Error|%v", err)
		return err
	}
	return nil
}

func (ci *Chapter) QueryChapterInfo(ctx context.Context, ds *redis.Client) (*Chapter, error) {
	key := BuildChapterInfo(*ci.Chid)
	result, err := ds.Get(ctx, key).Result()
	if nil != err {
		return nil, err
	}

	err = json.Unmarshal([]byte(result), ci)
	if nil != err {
		return nil, err
	}
	return ci, nil
}

func (ci *Chapter) QueryResourcList(ctx context.Context, ds *redis.Client, referFid string, limit int64) ([]string, error) {
	zrangeBy := redis.ZRangeBy{
		Min:    "0",
		Max:    strconv.Itoa(math.MaxInt64),
		Offset: 0,
		Count:  limit,
	}

	if len(referFid) > 0 {
		result, err := ds.ZScore(ctx, BuildChapterResourceList(*ci.Chid), referFid).Result()
		if nil != err {
			return nil, err
		}
		zrangeBy.Min = "(" + strconv.FormatInt(int64(result), 10)
	}

	result, err := ds.ZRangeByScore(ctx, BuildChapterResourceList(*ci.Chid), &zrangeBy).Result()
	if nil != err {
		return nil, err
	}

	return result, nil
}

func (cl *ChapterList) QueryChapterList(ctx context.Context, ds *redis.Client) (*ChapterList, error) {

	key := BuildSourceChapterList(cl.SourceId)

	zrangeBy := redis.ZRangeBy{
		Min:    "0",
		Max:    strconv.Itoa(math.MaxInt64),
		Count:  cl.Limit,
		Offset: 0,
	}

	if len(cl.ReferChid) > 0 {
		result, err := ds.ZScore(ctx, key, cl.ReferChid).Result()
		if nil != err {
			logx.GetLogger("study").Errorf("QueryChapterList|Get ReferChid Score Error|%v", err)
			return nil, err
		}

		zrangeBy.Min = "(" + strconv.FormatInt(int64(result), 10)
	}

	result, err := ds.ZRangeByScore(ctx, key, &zrangeBy).Result()
	if nil != err {
		logx.GetLogger("study").Errorf("QueryChapterList|Get ReferChid Score Error|%v", err)
		return nil, err
	}

	for _, chid := range result {
		chapter := Chapter{Chid: &chid}
		info, err := chapter.QueryChapterInfo(ctx, ds)
		if nil != err {
			logx.GetLogger("study").Errorf("QueryChapterList|Query Chapter Info Error|%v|%s", err, chid)
			return nil, err
		}
		cl.ChapterList = append(cl.ChapterList, *info)
	}

	return cl, nil
}

func (ch *Chapter) SaveChapter(ctx context.Context, ds *redis.Client) error {
	key := BuildChapterInfo(*ch.Chid)
	marshal, err := json.Marshal(ch)
	if err != nil {
		return err
	}
	_, err = ds.Set(ctx, key, string(marshal), 0).Result()
	if err != nil {
		return err
	}
	return nil
}
