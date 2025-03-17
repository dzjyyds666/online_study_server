package server

//type Class struct {
//	Cid        string `json:"cid" gorm:"cid;primaryKey"`
//	ClassName  string `json:"class_name" gorm:"class_name"`
//	ClassType  string `json:"class_type" gorm:"class_type"`
//	ClassDesc  string `json:"class_desc" gorm:"class_desc"`
//	IsComplete int    `json:"is_complete" gorm:"is_complete"`
//	IsDelete   int    `json:"is_delete" gorm:"is_delete"`
//	Owner      string `json:"owner" gorm:"owner"`
//	CreateTs   int64  `json:"create_ts" gorm:"create_ts"`
//	UpdateTs   int64  `json:"update_ts" gorm:"upload_ts"`
//}
//
//func (cl *Class) TableName() string {
//	return "class"
//}

var ClassStatus = struct {
	Complete   int
	UnComplete int

	Delete    int
	NotDelete int
}{
	UnComplete: 0,
	Complete:   1,

	Delete:    2,
	NotDelete: 3,
}
