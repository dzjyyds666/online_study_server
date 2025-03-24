package server

type ClassType struct {
	CTid     string `json:"ct_id" gorm:"ct_id;primaryKey"`
	CTname   string `json:"ct_name" gorm:"ct_name"`
	CTdesc   string `json:"ct_desc" gorm:"ct_desc"`
	CTstatus int    `json:"ct_status" gorm:"ct_status"`
}

func (ct *ClassType) TableName() string {
	return "class_type"
}

var CTStatus = struct {
	Delete int
	Normal int
}{
	Delete: 0,
	Normal: 1,
}
