package server

type UserInfo struct {
	Uid      string `gorm:"uid;primaryKeys" json:"uid"`
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

var UserStatus = struct {
	Active   int8
	Inactive int8
	Deleted  int8
}{
	Active:   0,
	Inactive: 1,
	Deleted:  2,
}

var UserGender = struct {
	Male    int8
	Female  int8
	UnKnown int8
}{
	Male:    0,
	Female:  1,
	UnKnown: 2,
}
