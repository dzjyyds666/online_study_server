package userHttpService

type UserInfo struct {
	Uid      string `gorm:"uid;primaryKey" json:"uid"`
	Username string `json:"username" gorm:"username;index:idx_username"`
	Account  string `json:"account" gorm:"account;index:idx_account;unique"`
	Password string `json:"password" gorm:"password"`
	Avatar   string `json:"avatar" gorm:"avatar"`
	Gender   int8   `json:"gender" gorm:"gender"`
	Role     int    `json:"role" gorm:"role"`
	Status   int8   `json:"status" gorm:"status"`
	CreateTs int64  `json:"create_ts" gorm:"create_ts"`
	UpdateTs int64  `json:"update_ts" gorm:"update_ts"`
}

var UserRole = struct {
	Admin   int
	Teacher int
	Student int
}{
	Admin:   3,
	Teacher: 2,
	Student: 1,
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
