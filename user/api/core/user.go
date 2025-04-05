package core

import "encoding/json"

type UserInfo struct {
	Uid      string `gorm:"uid;primaryKey" json:"uid"`
	Password string `json:"password" gorm:"password"`
	Avatar   string `json:"avatar" gorm:"avatar"`
	Role     int    `json:"role" gorm:"role"`
	Status   int8   `json:"status" gorm:"status"`
	Collage  string `json:"collage" gorm:"collage"` // 学院
	// 专业
	Major    string `json:"major" gorm:"major"`
	CreateTs int64  `json:"create_ts" gorm:"create_ts"`
	UpdateTs int64  `json:"update_ts" gorm:"update_ts"`
}

func (ui *UserInfo) WithUid(uid string) *UserInfo {
	ui.Uid = uid
	return ui
}

func (ui *UserInfo) WithPassword(password string) *UserInfo {
	ui.Password = password
	return ui
}

func (ui *UserInfo) WithAvatar(avatar string) *UserInfo {
	ui.Avatar = avatar
	return ui
}

func (ui *UserInfo) WithRole(role int) *UserInfo {
	ui.Role = role
	return ui
}

func (ui *UserInfo) WithStatus(status int8) *UserInfo {
	ui.Status = status
	return ui
}
func (ui *UserInfo) WithCollage(collage string) *UserInfo {
	ui.Collage = collage
	return ui
}
func (ui *UserInfo) WithMajor(major string) *UserInfo {
	ui.Major = major
	return ui
}
func (ui *UserInfo) WithCreateTs(ts int64) *UserInfo {
	ui.CreateTs = ts
	return ui
}
func (ui *UserInfo) WithUpdateTs(ts int64) *UserInfo {
	ui.UpdateTs = ts
	return ui
}

func (ui *UserInfo) Marshal() string {
	raw, _ := json.Marshal(ui)
	return string(raw)
}

func UnmarshalToUserInfo(data []byte) (*UserInfo, error) {
	var user UserInfo
	err := json.Unmarshal(data, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
