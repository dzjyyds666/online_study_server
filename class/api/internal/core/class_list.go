package core

import (
	"context"
	"encoding/json"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/redis/go-redis/v9"
	"math"
	"strconv"
)

type ClassList struct {
	ClassInfos []Class `json:"class_infos"`
	Limit      *int64  `json:"limit"`
	ReferCid   *string `json:"refer_cid"`
	Deleted    *bool   `json:"deleted"`
}

func (cl *ClassList) WithClassInfos(infos []Class) *ClassList {
	cl.ClassInfos = infos
	return cl
}

func (cl *ClassList) WithLimit(limit int64) *ClassList {
	cl.Limit = &limit
	return cl
}

func (cl *ClassList) WithReferCid(cid string) *ClassList {
	cl.ReferCid = &cid
	return cl
}

func (cl *ClassList) QueryClassList(ctx context.Context, uid string, ds *redis.Client) (ClassList, error) {
	classKey := BuildTeacherClassList(uid)
	zrangeBy := redis.ZRangeBy{
		Min:    "0",
		Max:    strconv.FormatInt(math.MaxInt64, 10),
		Offset: 0,
		Count:  *cl.Limit,
	}

	if cl.ReferCid != nil {
		score, err := ds.ZScore(ctx, classKey, *cl.ReferCid).Result()
		if err != nil {
			logx.GetLogger("study").Errorf("QueryClassList|Get ReferCid Score Error|%v", err)
			return ClassList{}, err
		}

		zrangeBy.Min = "(" + strconv.FormatInt(int64(score), 10)
	}

	classes, err := ds.ZRangeByScore(ctx, classKey, &zrangeBy).Result()
	if err != nil {
		logx.GetLogger("study").Errorf("QueryClassList|Get Class List Error|%v", err)
		return ClassList{}, err
	}

	// 获取当前课程的信息
	for _, class := range classes {
		var classInfo Class
		result, err := ds.Get(ctx, BuildClassInfo(class)).Result()
		if err != nil {
			logx.GetLogger("study").Errorf("QueryClassList|Get Class Info Error|%v", err)
			return ClassList{}, err
		}
		err = json.Unmarshal([]byte(result), &classInfo)
		if err != nil {
			logx.GetLogger("study").Errorf("QueryClassList|Unmarshal Class Info Error|%v", err)
			return ClassList{}, err
		}

		cl.ClassInfos = append(cl.ClassInfos, classInfo)
	}

	return *cl, nil
}
