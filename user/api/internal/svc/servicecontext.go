package svc

import (
	"github.com/redis/go-redis/v9"
	"github/dzjyyds666/online_study_server/user/api/internal/config"
	"github/dzjyyds666/online_study_server/user/api/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config config.Config
	Redis  *redis.Client
	Mysql  *gorm.DB
}

func NewServiceContext(c config.Config) *ServiceContext {

	redisClient := redis.NewClient(&redis.Options{
		Addr:     c.Redis.Host + ":" + c.Redis.Port,
		Password: c.Redis.Password,
		DB:       c.Redis.DB,
	})

	dsn := c.Mysql.User + ":" + c.Mysql.Password + "@tcp(" + c.Mysql.Host + ":" + c.Mysql.Port + ")/" + c.Mysql.DB + "?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	// 自动迁移
	err = db.AutoMigrate(&models.User{})
	if err != nil {
		panic(err)
	}

	return &ServiceContext{
		Config: c,
		Redis:  redisClient,
		Mysql:  db,
	}
}
