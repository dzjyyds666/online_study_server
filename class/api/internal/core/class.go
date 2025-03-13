package core

type ClassInfo struct {
	Cid       *string `json:"cid"`
	ClassName *string `json:"class_name"`
	ClassDesc *string `json:"class_desc"`
	ClassType *string `json:"class_type"`
	CreateTs  *int64  `json:"create_ts"`
	Teacher   *string `json:"teacher"`
	Archive   *bool   `json:"archive"`
}

func (ci *ClassInfo) WithCid(id string) *ClassInfo {
	ci.Cid = &id
	return ci
}

func (ci *ClassInfo) WithClassName(name string) *ClassInfo {
	ci.ClassName = &name
	return ci
}

func (ci *ClassInfo) WithClassDesc(desc string) *ClassInfo {
	ci.ClassDesc = &desc
	return ci
}

func (ci *ClassInfo) WithClassType(type_ string) *ClassInfo {
	ci.ClassType = &type_
	return ci
}

func (ci *ClassInfo) WithCreateTs(ts int64) *ClassInfo {
	ci.CreateTs = &ts
	return ci
}

func (ci *ClassInfo) WithTeacher(teacher string) *ClassInfo {
	ci.Teacher = &teacher
	return ci
}

func (ci *ClassInfo) WithArchive(archive bool) *ClassInfo {
	ci.Archive = &archive
	return ci
}
