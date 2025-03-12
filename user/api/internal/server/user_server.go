package server

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/dzjyyds666/opensource/common"
	"github.com/dzjyyds666/opensource/httpx"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/dzjyyds666/opensource/sdk"
	"github.com/labstack/echo"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
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

	err = mysql.AutoMigrate(&UserInfo{})
	if err != nil {
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
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	if err = bcrypt.CompareHashAndPassword([]byte(userInfo.Password), []byte(loginInfo.Password)); err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerLogin|Password not match|%s|%v", loginInfo.Account, err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Password Not Match",
		})
	}

	// 生成token
	tokenInfo := mymiddleware.Token{
		Uid:  userInfo.Uid,
		Role: userInfo.Role,
	}

	logx.GetLogger("OS_Server").Infof("HandlerLogin|Login Success|%s", common.ToStringWithoutError(tokenInfo))

	bytes, _ := json.Marshal(tokenInfo)
	token, err := sdk.CreateJwtToken(*config.GloableConfig.Jwt.Secretkey, bytes)
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerLogin|CreateJwtToken Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "login Error",
		})
	}

	err = us.redis.Set(ctx.Request().Context(), fmt.Sprintf(core.RedisTokenKey, tokenInfo.Uid), token, time.Duration(*config.GloableConfig.Jwt.Expire)*time.Second).Err()
	if nil != err {
		logx.GetLogger("OS_Server").Errorf("HandlerLogin|Token Set To Redis Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Token Set To Redis Error",
		})
	}

	logx.GetLogger("OS_Server").Infof("HandlerLogin|Login Success|%s", common.ToStringWithoutError(tokenInfo))
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, echo.Map{
		"token": token,
		"id":    userInfo.Uid,
		"role":  userInfo.Role,
	})
}

func (us *UserServer) HandleSignUp(ctx echo.Context) error {
	var signUpInfo login
	if err := ctx.Bind(&signUpInfo); err != nil {
		logx.GetLogger("OS_Server").Errorf("HandleSignUp|ctx.Bind err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	logx.GetLogger("OS_Server").Infof("HandleSignUp|SignUp Success|%s", common.ToStringWithoutError(signUpInfo))

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
		logx.GetLogger("OS_Server").Errorf("HandleSignUp|Create User Error|%v", err)
		//return ctx.JSON(http.StatusBadRequest,
		//	httpx.HttpResponse{StatusCode: httpx.HttpStatusCode.HttpBadRequest, Msg: "System failure"})
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Sign Up Error",
		})
	}

	logx.GetLogger("OS_Server").Infof("HandleSignUp|Create User Success|%s", common.ToStringWithoutError(userInfo))
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, userInfo)
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
		//return ctx.JSON(http.StatusBadRequest,
		//	httpx.HttpResponse{StatusCode: httpx.HttpStatusCode.HttpInternalError, Msg: "Send Message Error"})
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Send Message Error",
		})
	}

	// 把生成的验证码存入到redis中
	err := us.redis.SetNX(ctx.Request().Context(), fmt.Sprintf(core.RedisVerifyCodeKey, to), verfiyCode, time.Minute*5).Err()
	if nil != err {
		logx.GetLogger("OS_Server").Errorf("HandlerListUsers|VerifyCode Set To Redis|Error|%v", err)
		//return ctx.JSON(http.StatusBadRequest,
		//	httpx.HttpResponse{StatusCode: httpx.HttpStatusCode.HttpInternalError, Msg: "Send Message Error"})
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Send Message Error",
		})
	}

	//return ctx.JSON(http.StatusOK, httpx.HttpResponse{StatusCode: httpx.HttpStatusCode.HttpOK, Msg: "Send Message Success"})
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, echo.Map{
		"msg": "Send Message Success",
	})
}

func (us *UserServer) UpdateUserInfo(ctx echo.Context) error {
	var userInfo UserInfo
	if err := ctx.Bind(userInfo); err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerListUsers|ctx.Bind err:%v", err)
		//return ctx.JSON(http.StatusBadRequest,
		//	httpx.HttpResponse{StatusCode: httpx.HttpStatusCode.HttpBadRequest, Msg: "Params Invalid"})
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	if err := us.mysql.Where("uid = ?", userInfo.Uid).Updates(&userInfo).Error; err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerListUsers|Update User Info Error|%v", err)
		//return ctx.JSON(http.StatusBadRequest,
		//	httpx.HttpResponse{StatusCode: httpx.HttpStatusCode.HttpBadRequest, Msg: "System failure"})
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Update User Info Error",
		})
	}

	//return ctx.JSON(http.StatusOK, httpx.HttpResponse{StatusCode: httpx.HttpStatusCode.HttpOK, Msg: "Update User Info Success"})
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, echo.Map{
		"msg": "Update User Info Success",
	})
}

func (us *UserServer) HandlerQueryUserInfo(ctx echo.Context) error {
	uid := ctx.Get("uid").(string)
	if len(uid) <= 0 {
		logx.GetLogger("OS_Server").Errorf("HandlerListUsers|uid is empty")
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	// 数据库中查询用户信息
	var userInfo UserInfo
	err := us.mysql.First(&userInfo, "uid = ?", uid).Select("username", "account", "avatar", "role", "create_ts", "gender").Error
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerListUsers|err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Query User Info Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, userInfo)
}

type Page struct {
	Limit      int `query:"limit"`
	PageNumber int `query:"page_number"`
}

func (us *UserServer) HandlerListUsers(ctx echo.Context) error {
	var page Page
	err := ctx.Bind(&page)
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerListUsers|err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	offset := (page.PageNumber - 1) * page.Limit

	var userInfo []UserInfo
	if err := us.mysql.Limit(page.Limit).Offset(offset).Find(&userInfo).Error; err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerListUsers|err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Query User Info Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, userInfo)
}

func (us *UserServer) HandlerDeleteUser(ctx echo.Context) error {
	uids := make([]string, 0)
	if err := ctx.Bind(&uids); err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerListUsers|err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	for _, uid := range uids {
		if err := us.mysql.Delete(&UserInfo{}, "uid = ?", uid).Error; err != nil {
			logx.GetLogger("OS_Server").Errorf("HandlerListUsers|err:%v", err)
			return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
				"msg": "Delete User Info Error",
			})
		}
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, echo.Map{
		"msg": "Delete User Info Success",
	})
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
