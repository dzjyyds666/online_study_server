package core

import "fmt"

const (
	RedisUserCommentListKey    = "learnX:user:%s:comment:list"
	RedisArticleCommentListKey = "learnX:article:%s:comment:list"
)

func buildUserCommentListKey(uid string) string {
	return fmt.Sprintf(RedisUserCommentListKey, uid)
}

func buildArticleCommentListKey(articleId string) string {
	return fmt.Sprintf(RedisArticleCommentListKey, articleId)
}

type Comment struct {
	Id         string   `bson:"_id,omitempty" json:"id"`
	ArticleId  string   `bson:"article_id,omitempty" json:"article_id"`
	Author     string   `bson:"author,omitempty" json:"author"`
	Content    string   `bson:"content,omitempty" json:"content"`
	CreateTs   int64    `bson:"create_ts,omitempty" json:"create_ts"`
	Likes      int64    `bson:"likes,omitempty" json:"likes"`
	DisLikes   int64    `bson:"dislikes,omitempty" json:"dislikes"`
	ParentId   string   `bson:"parent_id,omitempty" json:"parent_id"`
	Replies    []string `bson:"replies,omitempty" json:"replies"`
	AuthorName string   `json:"author_name,omitempty"`
	Role       int64    `json:"role,omitempty"`
}

func (c *Comment) WithRole(role int64) *Comment {
	c.Role = role
	return c
}

func (c *Comment) WithId(id string) *Comment {
	c.Id = id
	return c
}

func (c *Comment) WithArticleId(articleId string) *Comment {
	c.ArticleId = articleId
	return c
}
func (c *Comment) WithAuthor(author string) *Comment {
	c.Author = author
	return c
}

func (c *Comment) WithContent(content string) *Comment {
	c.Content = content
	return c
}

func (c *Comment) WithCreateTs(ts int64) *Comment {
	c.CreateTs = ts
	return c
}

func (c *Comment) WithLikes(likes int64) *Comment {
	c.Likes = likes
	return c
}

func (c *Comment) WithDisLikes(dislikes int64) *Comment {
	c.DisLikes = dislikes
	return c
}

func (c *Comment) WithParentId(parentId string) *Comment {
	c.ParentId = parentId
	return c
}
func (c *Comment) WithReplies(replies []string) *Comment {
	c.Replies = replies
	return c
}

type ListComment struct {
	List      []*Comment `json:"list"`
	Uid       string     `json:"uid,omitempty"`
	ArticleId string     `json:"article_id,omitempty"`
}
