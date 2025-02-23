package svc

import (
	"github.com/dzjyyds666/online_study_server/user/api/internal/config"
	"github.com/dzjyyds666/online_study_server/user/api/internal/models"
	"github.com/redis/go-redis/v9"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strconv"
)

type ServiceContext struct {
	Config config.Config
	Redis  *redis.Client
	Mysql  *gorm.DB
}

func NewServiceContext(c config.Config) *ServiceContext {

	redisClient := redis.NewClient(&redis.Options{
		Addr:     c.Redis.Host + ":" + strconv.Itoa(c.Redis.Port),
		Password: c.Redis.Password,
		DB:       c.Redis.DB,
	})

	dsn := c.Mysql.UserName + ":" + c.Mysql.Password + "@tcp(" + c.Mysql.Host + ":" + strconv.Itoa(c.Mysql.Port) + ")/" + c.Mysql.DB + "?charset=utf8mb4&parseTime=True&loc=Local"

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
