package core

import (
	"encoding/json"
)

type Class struct {
	Cid            *string   `json:"cid,omitempty" bson:"_id,omitempty"`
	ClassName      *string   `json:"class_name,omitempty" bson:"class_name,omitempty"`
	ClassDesc      *string   `json:"class_desc,omitempty" bson:"class_desc,omitempty"`
	ClassType      *string   `json:"class_type,omitempty" bson:"class_type,omitempty"`
	CreateTs       *int64    `json:"create_ts,omitempty" bson:"create_ts,omitempty"`
	Teacher        *string   `json:"teacher,omitempty" bson:"teacher,omitempty"`
	Archive        *bool     `json:"archive,omitempty" bson:"archive,omitempty"`
	Deleted        *bool     `json:"deleted,omitempty" bson:"deleted,omitempty"`
	StudyClassName *string   `json:"study_class_name,omitempty" bson:"study_class_name,omitempty"`
	ChapterList    []Chapter `json:"chapter_list,omitempty" bson:"chapter_list,omitempty"`

	Cover           *string `json:"cover,omitempty" bson:"cover,omitempty"`
	ClassScore      *string `json:"class_score,omitempty" bson:"class_score,omitempty"`             // 学分
	ClassTime       *string `json:"class_time,omitempty" bson:"class_time,omitempty"`               // 学时
	ClassCollege    *string `json:"class_college,omitempty" bson:"class_college,omitempty"`         // 学院
	ClassSchoolTerm *string `json:"class_school_term,omitempty" bson:"class_school_term,omitempty"` // 学期
	ClassOutline    *string `json:"class_outline,omitempty" bson:"class_outline"`                   // 课程大纲
}

func (ci *Class) WithCid(id string) *Class {
	ci.Cid = &id
	return ci
}

func (ci *Class) WithClassName(name string) *Class {
	ci.ClassName = &name
	return ci
}

func (ci *Class) WithCover(cover string) *Class {
	ci.Cover = &cover
	return ci
}

func (ci *Class) WithClassDesc(desc string) *Class {
	ci.ClassDesc = &desc
	return ci
}

func (ci *Class) WithClassType(type_ string) *Class {
	ci.ClassType = &type_
	return ci
}

func (ci *Class) WithCreateTs(ts int64) *Class {
	ci.CreateTs = &ts
	return ci
}

func (ci *Class) WithTeacher(teacher string) *Class {
	ci.Teacher = &teacher
	return ci
}

func (ci *Class) WithArchive(archive bool) *Class {
	ci.Archive = &archive
	return ci
}

func (ci *Class) WithDeleted(deleted bool) *Class {
	ci.Deleted = &deleted
	return ci
}

func (ci *Class) WithClassScore(score string) *Class {
	ci.ClassScore = &score
	return ci
}

func (ci *Class) WithClassTime(time string) *Class {
	ci.ClassTime = &time
	return ci
}

func (ci *Class) WithClassCollege(college string) *Class {
	ci.ClassCollege = &college
	return ci
}

func (ci *Class) WithClassSchoolTerm(term string) *Class {
	ci.ClassSchoolTerm = &term
	return ci
}

func (ci *Class) WithClassOutline(outline string) *Class {
	ci.ClassOutline = &outline
	return ci
}

func (ci *Class) WithStudyClass(name string) *Class {
	ci.StudyClassName = &name
	return ci
}

func (ci *Class) ToJsonWithoutErr() string {
	marshal, _ := json.Marshal(ci)
	return string(marshal)
}

func (ci *Class) IsDeleted() bool {
	return *ci.Deleted
}

func UnmarshalToClass(data []byte) (*Class, error) {
	var ci Class
	err := json.Unmarshal(data, &ci)
	if err != nil {
		return nil, err
	}
	return &ci, nil
}
