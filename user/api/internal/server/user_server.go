package server

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/dzjyyds666/opensource/httpx"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/labstack/echo"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
	"user/api/config"
	"user/api/internal/core"
	mymiddleware "user/api/internal/middleware"
)

type UserServer struct {
	redis *redis.Client  // redis
	mysql *gorm.DB       // mysql数据库
	email *gomail.Dialer // 邮件
}

func NewUserServer() (*UserServer, error) {
	// 初始化redis和mysql
	client := redis.NewClient(&redis.Options{
		Addr:     *config.GloableConfig.Redis.Host + ":" + strconv.Itoa(*config.GloableConfig.Redis.Port),
		Username: *config.GloableConfig.Redis.Username,
		Password: *config.GloableConfig.Redis.Password,
		DB:       *config.GloableConfig.Redis.DB,
	})

	dialer := gomail.NewDialer(
		*config.GloableConfig.Email.Host,
		*config.GloableConfig.Email.Port,
		*config.GloableConfig.Email.User,
		*config.GloableConfig.Email.Password)

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
		email: dialer,
	}
	return userServer, nil
}

type login struct {
	Account  string `json:"account"`
	Password string `json:"password"`
}

// todo 邮箱验证码登录

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

	err = us.redis.Set(ctx.Request().Context(), core.RedisTokenKey, string(marshal), time.Duration(*config.GloableConfig.Jwt.Expire)*time.Second).Err()
	if nil != err {
		logx.GetLogger("OS_Server").Errorf("HandlerLogin|Token Set To Redis Error|%v", err)
		return ctx.JSON(http.StatusBadRequest,
			httpx.HttpResponse{StatusCode: httpx.HttpStatusCode.HttpInternalError, Msg: "Token Set To Redis Error"})
	}

	return ctx.JSON(http.StatusOK, httpx.HttpResponse{StatusCode: httpx.HttpStatusCode.HttpOK, Msg: "Login Success", Data: string(marshal)})
}

type signUp struct {
	Account    string `json:"account"`
	Password   string `json:"password"`
	VerifyCode string `json:"verify_code"`
}

func (us *UserServer) SignUp(ctx echo.Context) error {
	var signUpInfo signUp
	if err := ctx.Bind(&signUpInfo); err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerLogin|ctx.Bind err:%v", err)
		return ctx.JSON(http.StatusBadRequest,
			httpx.HttpResponse{StatusCode: httpx.HttpStatusCode.HttpBadRequest, Msg: "Params Invalid"})
	}

	// 从redis中获取到邮箱验证码
	result, err := us.redis.Get(ctx.Request().Context(), core.RedisVerifyCodeKey).Result()
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerListUsers|Get Verify Code From Redis Error|%v", err)
		return ctx.JSON(http.StatusBadRequest,
			httpx.HttpResponse{StatusCode: httpx.HttpStatusCode.HttpBadRequest, Msg: "System failure"})
	}

	if result != signUpInfo.VerifyCode {
		logx.GetLogger("OS_Server").Errorf("HandlerListUsers|Verify Code Not Match|%v", err)
		return ctx.JSON(http.StatusBadRequest,
			httpx.HttpResponse{StatusCode: httpx.HttpStatusCode.HttpBadRequest, Msg: "Verify Code Not Match"})
	}

	var userInfo UserInfo
	// 生成uid
	userInfo.Uid = "LX_" + GenerateRandomString(10)
	userInfo.Role = UserRole.Student
	userInfo.Username = "LearnX-" + GenerateRandomString(5)
	userInfo.Account = signUpInfo.Account
	password, err := bcrypt.GenerateFromPassword([]byte(signUpInfo.Password), bcrypt.DefaultCost)
	userInfo.Password = string(password)
	userInfo.CreateTs = time.Now().Unix()
	userInfo.UpdateTs = time.Now().Unix()
	userInfo.Status = UserStatus.Active
	userInfo.Gender = UserGender.UnKnown

	// 插入数据库
	if err = us.mysql.Create(&userInfo).Error; err != nil {
		logx.GetLogger("OS_Server").Errorf("SendMessage|Create User Error|%v", err)
		return ctx.JSON(http.StatusBadRequest,
			httpx.HttpResponse{StatusCode: httpx.HttpStatusCode.HttpBadRequest, Msg: "System failure"})
	}

	return ctx.JSON(http.StatusOK, httpx.HttpResponse{StatusCode: httpx.HttpStatusCode.HttpOK, Msg: "SignUp Success"})
}

func (us *UserServer) SendMessage(ctx echo.Context) error {
	to := ctx.Param("email")
	subject := ctx.Param("subject")

	htmlTemplate := `<!DOCTYPE html>
<html lang="zh">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>LearnX</title>
    <style>
        body { font-family: Arial, sans-serif; background-color: #f4f4f4; text-align: center; padding: 20px; }
        .email-container { background: white; padding: 20px; border-radius: 10px; box-shadow: 0 0 10px rgba(0, 0, 0, 0.1); display: inline-block; max-width: 400px; width: 100%; }
        .verification-code { font-size: 24px; font-weight: bold; color: #007bff; margin: 20px 0; }
        .footer { margin-top: 15px; font-size: 14px; color: #666; }
    </style>
</head>
<body>
    <div class="email-container">
        <h2>您的验证码</h2>
        <p>请使用以下验证码完成验证：</p>
        <p class="verification-code">%s</p>
		<p class="footer">验证码5分钟之内有效，请及时完成验证。</p>
        <p class="footer">如果您没有请求此验证码，请忽略此邮件。</p>
    </div>
</body>
</html>`

	verfiyCode := GenerateRandomString(6)

	htmlTemplate = fmt.Sprintf(htmlTemplate, verfiyCode)

	message := gomail.NewMessage()
	message.SetHeader("From", message.FormatAddress(*config.GloableConfig.Email.Sender, *config.GloableConfig.Email.Alias))
	message.SetHeader("To", to)
	message.SetHeader("Subject", subject)
	message.SetBody("text/html", htmlTemplate)

	if err := us.email.DialAndSend(message); err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerListUsers|SendMessage|Error|%v", err)
		return ctx.JSON(http.StatusBadRequest,
			httpx.HttpResponse{StatusCode: httpx.HttpStatusCode.HttpInternalError, Msg: "Send Message Error"})
	}

	// 把生成的验证码存入到redis中
	err := us.redis.SetNX(ctx.Request().Context(), fmt.Sprintf(core.RedisVerifyCodeKey, to), verfiyCode, time.Minute*5).Err()
	if nil != err {
		logx.GetLogger("OS_Server").Errorf("HandlerListUsers|VerifyCode Set To Redis|Error|%v", err)
		return ctx.JSON(http.StatusBadRequest,
			httpx.HttpResponse{StatusCode: httpx.HttpStatusCode.HttpInternalError, Msg: "Send Message Error"})
	}

	return ctx.JSON(http.StatusOK, httpx.HttpResponse{StatusCode: httpx.HttpStatusCode.HttpOK, Msg: "Send Message Success"})
}

func (us *UserServer) UpdateUserInfo(ctx echo.Context) error {

	var userInfo UserInfo
	if err := ctx.Bind(userInfo); err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerListUsers|ctx.Bind err:%v", err)
		return ctx.JSON(http.StatusBadRequest,
			httpx.HttpResponse{StatusCode: httpx.HttpStatusCode.HttpBadRequest, Msg: "Params Invalid"})
	}

	if err := us.mysql.Where("uid = ?", userInfo.Uid).Updates(&userInfo).Error; err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerListUsers|Update User Info Error|%v", err)
		return ctx.JSON(http.StatusBadRequest,
			httpx.HttpResponse{StatusCode: httpx.HttpStatusCode.HttpBadRequest, Msg: "System failure"})
	}

	return ctx.JSON(http.StatusOK, httpx.HttpResponse{StatusCode: httpx.HttpStatusCode.HttpOK, Msg: "Update User Info Success"})
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
