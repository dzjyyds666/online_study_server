package models

// User represents the user table in the database.
type User struct {
	UserID        string `gorm:"primaryKey;column:user_id;type:varchar(16);not null" json:"user_id"`
	NickName      string `gorm:"column:nick_name;type:varchar(16);not null" json:"nick_name"`
	Avatar        string `gorm:"column:avatar;type:varchar(64);not null" json:"avatar"`
	Gender        string `gorm:"column:gender;type:varchar(1);not null" json:"gender"`
	Email         string `gorm:"column:email;type:varchar(32);unique;not null" json:"email"`
	Password      string `gorm:"column:password;type:varchar(64);not null" json:"password"`
	CreateTS      int    `gorm:"column:create_ts;type:int;not null" json:"create_ts"`
	Role          string `gorm:"column:role;type:varchar(1);not null" json:"role"`
	Status        string `gorm:"column:status;type:varchar(1);not null" json:"status"`
	LastLoginTime int    `gorm:"column:last_login_time;type:int;not null" json:"last_login_time"`
	LastLoginIP   string `gorm:"column:last_login_ip;type:varchar(32);not null" json:"last_login_ip"`
}
