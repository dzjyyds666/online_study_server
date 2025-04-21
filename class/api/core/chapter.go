package core

import (
	"encoding/json"
)

// 章节
type Chapter struct {
	Chid         *string    `json:"chid,omitempty"`
	ChapterName  *string    `json:"chapter_name,omitempty"`
	SourceId     *string    `json:"source_id,omitempty"`
	ResourceList []Resource `json:"resource_list,omitempty"`
}

type ChapterList struct {
	SourceId    string    `json:"source_id"`
	ReferChid   string    `json:"refer_chid"`
	Limit       int64     `json:"limit"`
	ChapterList []Chapter `json:"chapter_list"`
}

func (ch *Chapter) WithChid(id string) *Chapter {
	ch.Chid = &id
	return ch
}

func (ch *Chapter) WithChapterName(name string) *Chapter {
	ch.ChapterName = &name
	return ch
}

func (ch *Chapter) WithSourceId(id string) *Chapter {
	ch.SourceId = &id
	return ch
}

func (ch *Chapter) Marshal() string {
	marshal, _ := json.Marshal(ch)
	return string(marshal)
}
