package server

import (
	"crypto/rand"
	"encoding/json"
	"github.com/dzjyyds666/opensource/httpx"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/labstack/echo"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
	"user/config"
	mymiddleware "user/internal/middleware"
)

const redisTokenKey = "user:login:token"

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

type login struct {
	Account  string `json:"account"`
	Password string `json:"password"`
}

func (us *UserServer) HandlerLogin(ctx echo.Context) error {
	var loginInfo login
	if err := ctx.Bind(&loginInfo); err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerLogin|ctx.Bind err:%v", err)
		return err
	}

	// 从数据库中查询用户信息
	var userInfo UserInfo
	err := us.mysql.Model(&UserInfo{}).Where("account = ?", loginInfo.Account).First(&userInfo).Error
	if nil != err {
		logx.GetLogger("OS_Server").Errorf("HandlerLogin|Can not Find User|%s|%v", loginInfo.Account, err)
		return ctx.JSON(http.StatusBadRequest, httpx.HttpResponse{StatusCode: httpx.HttpStatusCode.HttpBadRequest, Msg: "User Not Exits"})
	}

	if err = bcrypt.CompareHashAndPassword([]byte(userInfo.Password), []byte(loginInfo.Password)); err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerLogin|Password not match|%s|%v", loginInfo.Account, err)
		return ctx.JSON(http.StatusBadRequest, httpx.HttpResponse{StatusCode: httpx.HttpStatusCode.HttpBadRequest, Msg: "Password Not Match"})
	}

	// 生成token
	tokenInfo := mymiddleware.Token{
		Uid:  userInfo.Uid,
		Role: userInfo.Role,
	}

	marshal, err := json.Marshal(&tokenInfo)
	if nil != err {
		logx.GetLogger("OS_Server").Errorf("HandlerLogin|JSON Marshal Error|%v", err)
		return ctx.JSON(http.StatusBadRequest, httpx.HttpResponse{StatusCode: httpx.HttpStatusCode.HttpInternalError, Msg: "JSON Marshal Error"})
	}

	err = us.redis.Set(ctx.Request().Context(), redisTokenKey, string(marshal), time.Duration(*config.GloableConfig.Jwt.Expire)*time.Second).Err()
	if nil != err {
		logx.GetLogger("OS_Server").Errorf("HandlerLogin|Token Set To Redis Error|%v", err)
		return ctx.JSON(http.StatusBadRequest,
			httpx.HttpResponse{StatusCode: httpx.HttpStatusCode.HttpInternalError, Msg: "Token Set To Redis Error"})
	}

	return ctx.JSON(http.StatusOK, httpx.HttpResponse{StatusCode: httpx.HttpStatusCode.HttpOK, Msg: "Login Success", Data: string(marshal)})
}

func (us *UserServer) SignUp(ctx echo.Context) error {
	var signUpInfo UserInfo
	if err := ctx.Bind(&signUpInfo); err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerLogin|ctx.Bind err:%v", err)
		return ctx.JSON(http.StatusBadRequest,
			httpx.HttpResponse{StatusCode: httpx.HttpStatusCode.HttpBadRequest, Msg: "Params Invalid"})
	}

	// 生成uid
	signUpInfo.Uid = "learnX-" + GenerateRandomString(10)

	return nil
}

func (us *UserServer) HandlerListUsers(ctx echo.Context) error {
	return ctx.String(200, "hello world")
}

func GenerateRandomString(length int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return ""
	}
	for i, b := range bytes {
		bytes[i] = letters[b%byte(len(letters))]
	}
	return string(bytes)
}
