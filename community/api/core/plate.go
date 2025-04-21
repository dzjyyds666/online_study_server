package core

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
