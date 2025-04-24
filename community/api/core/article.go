package core

import (
	"fmt"
	"net/url"
)

type tag string

// 特殊前缀
func (t tag) S() string {
	return "learnX-" + string(t)
}

var Tags = struct {
}{}

const (
	RedisPlateArticleListKey      = "class:plate:%s:article:list"       // 板块文章列表
	RedisPlateArticleAuditListKey = "class:plate:%s:article:audit:list" // 待审核列表
)

func buildPlateArticleAuditListKey(plateId string) string {
	return fmt.Sprintf(RedisPlateArticleAuditListKey, plateId)
}

func buildPlateArticleListKey(plateId string) string {
	return fmt.Sprintf(RedisPlateArticleListKey, plateId)
}

type Article struct {
	Id          string     `json:"id" bson:"_id"`
	Title       string     `json:"title" bson:"title"`
	Content     string     `json:"content" bson:"content"`
	Attachments []string   `bson:"attachments" json:"attachments"`
	Author      string     `json:"author" bson:"author"`
	CreateTs    int64      `json:"create" bson:"create_ts"`
	UpdateTs    int64      `json:"update" bson:"update_ts"`
	PlateId     string     `json:"plate_id" bson:"plate_id"`
	Tags        url.Values `json:"tags" bson:"tags"`
}

func (a *Article) WithId(id string) *Article {
	a.Id = id
	return a
}

func (a *Article) WithTitle(title string) *Article {
	a.Title = title
	return a
}

func (a *Article) WithContent(content string) *Article {
	a.Content = content
	return a
}

func (a *Article) WithAttachments(attachments []string) *Article {
	a.Attachments = attachments
	return a
}

func (a *Article) WithAuthor(author string) *Article {
	a.Author = author
	return a
}

func (a *Article) WithCreateTs(ts int64) *Article {
	a.CreateTs = ts
	return a
}

func (a *Article) WithUpdateTs(ts int64) *Article {
	a.UpdateTs = ts
	return a
}

func (a *Article) WithPlateId(plateId string) *Article {
	a.PlateId = plateId
	return a
}
