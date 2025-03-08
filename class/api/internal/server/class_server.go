package server

import (
	"class/api/config"
	"context"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/labstack/echo"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strconv"
)

type ClassServer struct {
	redis *redis.Client
	mysql *gorm.DB
	mongo *mongo.Database
}

func NewClassServer() (*ClassServer, error) {

	// 连接redis
	dsClient := redis.NewClient(&redis.Options{
		Addr:     *config.GloableConfig.Redis.Host + ":" + strconv.Itoa(*config.GloableConfig.Redis.Port),
		Username: *config.GloableConfig.Redis.Username,
		Password: *config.GloableConfig.Redis.Password,
		DB:       *config.GloableConfig.Redis.DB,
	})

	// 连接mysql
	dsn := *config.GloableConfig.Mysql.Username + ":" + *config.GloableConfig.Mysql.Password + "@tcp(" + *config.GloableConfig.Mysql.Host + ":" + strconv.Itoa(*config.GloableConfig.Mysql.Port) + ")/" + *config.GloableConfig.Mysql.Database + "?charset=utf8mb4&parseTime=True&loc=Local"
	msClient, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("NewClassServer|gorm.Open err:%v", err)
		return nil, err
	}

	msClient.AutoMigrate(&ClassType{}, &Class{})

	mongodsn := "mongodb://" + *config.GloableConfig.Mongo.Username + ":" + *config.GloableConfig.Mongo.Password + "@ip:" + *config.GloableConfig.Mongo.Host + "/" + strconv.Itoa(*config.GloableConfig.Mongo.Port)
	// 连接mongo
	mgClient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongodsn))
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("NewClassServer|mongo.Connect err:%v", err)
		return nil, err
	}
	database := mgClient.Database(*config.GloableConfig.Mongo.DB)
	return &ClassServer{
		redis: dsClient,
		mysql: msClient,
		mongo: database,
	}, nil
}

func (cls *ClassServer) HandlerCreateClass(ctx echo.Context) error {

	return nil
}
