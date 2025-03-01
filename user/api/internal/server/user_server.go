package server

import (
	"github.com/dzjyyds666/opensource/logx"
	"github.com/labstack/echo"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strconv"
	"user/config"
)

type UserServer struct {
	redis *redis.Client // redis
	mysql *gorm.DB      // mysql数据库
}

func NewUserServer() (*UserServer, error) {
	// 初始化redis和mysql
	client := redis.NewClient(&redis.Options{
		Addr:     *config.GloableConfig.Redis.Host + ":" + strconv.Itoa(*config.GloableConfig.Redis.Port),
		Username: *config.GloableConfig.Redis.Username,
		Password: *config.GloableConfig.Redis.Password,
		DB:       *config.GloableConfig.Redis.DB,
	})

	// gorm连接mysql
	dsn := *config.GloableConfig.Mysql.Username + ":" + *config.GloableConfig.Mysql.Password + "@tcp(" + *config.GloableConfig.Mysql.Host + ":" + strconv.Itoa(*config.GloableConfig.Mysql.Port) + ")/" + *config.GloableConfig.Mysql.Database + "?charset=utf8mb4&parseTime=True&loc=Local"
	mysql, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if nil != err {
		logx.GetLogger("OS_Server").Errorf("NewUserServer|gorm.Open err:%v", err)
		return nil, err
	}

	userServer := &UserServer{
		redis: client,
		mysql: mysql,
	}
	return userServer, nil
}

func (us *UserServer) HandlerListUsers(ctx echo.Context) error {
	return ctx.String(200, "hello world")
}

func (us *UserServer) HandlerLogin(ctx echo.Context) error {
	return ctx.String(200, "hello world")
}
