package server

type UserInfo struct {
	Uid      string `gorm:"uid" json:"uid"`
	Username string `json:"username" gorm:"username"`
	Account  string `json:"account" gorm:"account"`
	Password string `json:"password" gorm:"password"`
	Avatar   string `json:"avatar" gorm:"avatar"`
	Gender   int8   `json:"gender" gorm:"gender"`
	Role     string `json:"role" gorm:"role"`
	Status   int8   `json:"status" gorm:"status"`
	CreateTs int64  `json:"create_ts" gorm:"create_ts"`
	UpdateTs int64  `json:"update_ts" gorm:"update_ts"`
}

var UserRole = struct {
	Admin   string
	Teacher string
	Student string
}{
	Admin:   "admin",
	Teacher: "teacher",
	Student: "student",
}
