package core

type Class struct {
	Cid          *string      `json:"cid"`
	ClassName    *string      `json:"class_name"`
	ClassDesc    *string      `json:"class_desc"`
	ClassType    *string      `json:"class_type"`
	CreateTs     *int64       `json:"create_ts"`
	Teacher      *string      `json:"teacher"`
	Archive      *bool        `json:"archive"`
	Deleted      *bool        `json:"deleted"`
	ChapterLists []Chapter    `json:"chapter_lists"` // 章节
	StudyClass   []StudyClass `json:"study_class"`   // 教学班
}

func (ci *Class) WithCid(id string) *Class {
	ci.Cid = &id
	return ci
}

func (ci *Class) WithClassName(name string) *Class {
	ci.ClassName = &name
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
