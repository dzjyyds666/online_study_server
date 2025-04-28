package core

import (
	"fmt"
)

const (
	RedisUserArticleListKey  = "user:%s:article:list"
	RedisArticleAuditListKey = "article:audit:list"
	RedisPlateArticleListKey = "plate:%s:article:list"
	RedisArticleList         = "article:list"
)

func buildArticleAuditListKey() string {
	return RedisArticleAuditListKey
}

func buildUserArticleListKey(userId string) string {
	return fmt.Sprintf(RedisUserArticleListKey, userId)
}

func buildPlateArticleListKey(plateId string) string {
	return fmt.Sprintf(RedisPlateArticleListKey, plateId)
}

func buildArticleListKey() string {
	return RedisArticleList
}

type Status string

var ArticleStatuses = struct {
	Published Status
	Audit     Status
	Illegal   Status
}{
	Published: "published", // 已发布
	Audit:     "audit",     // 审核中
	Illegal:   "illegal",   // 违规
}

type Article struct {
	Id          string   `json:"id" bson:"_id"`
	Title       string   `json:"title" bson:"title"`
	Content     string   `json:"content" bson:"content"`
	Attachments []string `bson:"attachments" json:"attachments"`
	Author      string   `json:"author" bson:"author"`
	CreateTs    int64    `json:"create_ts" bson:"create_ts"`
	UpdateTs    int64    `json:"update_ts" bson:"update_ts"`
	PlateId     string   `json:"plate_id" bson:"plate_id"`
	Status      Status   `json:"status" bson:"status"`
}

func (a *Article) WithStatus(status Status) *Article {
	a.Status = status
	return a
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

type ListArticle struct {
	List    []*Article `json:"list"`
	PlateId string     `json:"plate_id,omitempty"`
	Uid     string     `json:"uid,omitempty"`
	Audit   bool       `json:"audit,omitempty"`
	New     bool       `json:"new,omitempty"`
}
