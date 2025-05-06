package core

import (
	"encoding/json"
	"errors"
)

const (
	RedisPlateListKey = "class:plate:list" // 所有的板块的列表
)

func buildPlateListKey() string {
	return RedisPlateListKey
}

var (
	ErrNoMatchData = errors.New("no match data")
)

// 板块，社区中以板块为主要组成单位，板块中包含文章，文章中包含评论
type Plate struct {
	Id              string `json:"id,omitempty" bson:"_id,omitempty"`
	Name            string `json:"name,omitempty" bson:"name,omitempty"`
	Description     string `json:"description,omitempty" bson:"description,omitempty"`
	CreateTime      int64  `json:"create_time,omitempty" bson:"create_time,omitempty"`
	ArticleNumber   int64  `json:"article_number,omitempty" bson:"article_number,omitempty"`     // 板块中包含的文章数量
	SubscribeNumber int64  `json:"subscribe_number,omitempty" bson:"subscribe_number,omitempty"` // 板块中订阅的数量
}

func (p *Plate) WithId(id string) *Plate {
	p.Id = id
	return p
}

func (p *Plate) WithName(name string) *Plate {
	p.Name = name
	return p
}

func (p *Plate) WithDescription(description string) *Plate {
	p.Description = description
	return p
}

func (p *Plate) WithCreateTime(createTime int64) *Plate {
	p.CreateTime = createTime
	return p
}

func (p *Plate) WithArticleNumber(articleNumber int64) *Plate {
	p.ArticleNumber = articleNumber
	return p
}

func (p *Plate) WithSubscribeNumber(subscribeNumber int64) *Plate {
	p.SubscribeNumber = subscribeNumber
	return p
}

func (p *Plate) Marshal() (string, error) {
	marshal, err := json.Marshal(p)
	return string(marshal), err
}

type ListPlate struct {
	List       []*Plate `json:"list,omitempty"`
	PageSize   int64    `json:"page_size"`
	PageNumber int64    `json:"page_number"`
	Total      int64    `json:"total"`
}
