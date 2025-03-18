package core

import (
	"context"
	"encoding/json"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/redis/go-redis/v9"
	"math"
	"strconv"
	"time"
)

type StudyClass struct {
	SCid          string `json:"sc_id"`
	ClassName     string `json:"class_name"`
	StudentNumber int64  `json:"student_number"`
	Cid           string `json:"cid"`
	Tid           string `json:"tid"`
}

func (s *StudyClass) WithSCid(id string) *StudyClass {
	s.SCid = id
	return s
}

func (s *StudyClass) WithClassName(name string) *StudyClass {
	s.ClassName = name
	return s
}

func (s *StudyClass) WithStudentNumber(number int64) *StudyClass {
	s.StudentNumber = number
	return s
}

// 把教学班加入老师的教学班列表，课程对应的教学班列表
func (sc *StudyClass) CreateStudyClass(ctx context.Context, ds *redis.Client) error {
	teacherKey := BuildStudyClassTeacherKey(sc.Tid)
	err := ds.ZAdd(ctx, teacherKey, redis.Z{
		Member: sc.SCid,
		Score:  float64(time.Now().Unix()),
	}).Err()
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("CreateStudyClass|Add SCid To Teacher List Error|%v", err)
		return err
	}

	classKey := BuildStudyClassClassKey(sc.Cid)
	err = ds.ZAdd(ctx, classKey, redis.Z{
		Member: sc.SCid,
		Score:  float64(time.Now().Unix()),
	}).Err()
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("CreateStudyClass|Add SCid To Class List Error|%v", err)
		return err
	}

	// 添加教学班的info
	infoKey := BuildStudyClassInfoKey(sc.SCid)
	marshal, err := json.Marshal(sc)
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("CreateStudyClass|Marshal StudyClass Error|%v", err)
		return err
	}

	err = ds.Set(ctx, infoKey, string(marshal), 0).Err()
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("CreateStudyClass|Set StudyClass Info Error|%v", err)
		return err
	}

	return nil
}

type StudyClassList struct {
	StudyClasses []StudyClass `json:"study_classes"`
	ReferSCid    string       `json:"refer_scid"`
	Limit        int64        `json:"limit"`
	Tid          string       `json:"tid"`
}

func (sc *StudyClass) QueryStudyClassList(ctx context.Context, ds *redis.Client, list *StudyClassList) (*StudyClassList, error) {

	zrangeBy := &redis.ZRangeBy{
		Min:    "0",
		Max:    strconv.FormatInt(math.MaxInt64, 10),
		Offset: 0,
		Count:  list.Limit,
	}

	if len(list.ReferSCid) > 0 {
		score, err := ds.ZScore(ctx, BuildStudyClassTeacherKey(list.Tid), list.ReferSCid).Result()
		if err != nil {
			logx.GetLogger("OS_Server").Errorf("QueryStudyClassList|Get ReferSCid Score Error|%v", err)
			return nil, err
		}
		zrangeBy.Min = "(" + strconv.FormatInt(int64(score), 10)
	}

	scids, err := ds.ZRangeByScore(ctx, BuildStudyClassClassKey(list.Tid), zrangeBy).Result()
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("QueryStudyClassList|Get SCid List Error|%v", err)
		return nil, err
	}

	for _, scid := range scids {
		infoKey := BuildStudyClassInfoKey(scid)
		result, err := ds.Get(ctx, infoKey).Result()
		if err != nil {
			logx.GetLogger("OS_Server").Errorf("QueryStudyClassList|Get SCid Info Error|%v", err)
			return nil, err
		}

		var studyClass StudyClass
		if err := json.Unmarshal([]byte(result), &studyClass); err != nil {
			logx.GetLogger("OS_Server").Errorf("QueryStudyClassList|Unmarshal StudyClass Info Error|%v", err)
			return nil, err
		}
		list.StudyClasses = append(list.StudyClasses, studyClass)
	}

	return list, nil
}
