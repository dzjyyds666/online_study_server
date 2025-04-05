package userHttpService

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"github.com/dzjyyds666/opensource/common"
	"github.com/dzjyyds666/opensource/httpx"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/labstack/echo"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
	"gorm.io/gorm"
	"time"
	core2 "user/api/core"
)

type UserService struct {
	ctx        context.Context
	userServer *core2.UserServer
}

func NewUserService(ctx context.Context, client *redis.Client, mysql *gorm.DB, dialer *gomail.Dialer) (*UserService, error) {
	err := mysql.AutoMigrate(&core2.UserInfo{})
	if err != nil {
		logx.GetLogger("study").Errorf("UserService|StartError|AutoMigrate|err:%v", err)
		return nil, err
	}
	userServer := &UserService{
		ctx:        ctx,
		userServer: core2.NewUserServer(ctx, client, mysql),
	}
	return userServer, nil
}

func (us *UserService) HandlerLogin(ctx echo.Context) error {
	var user core2.UserInfo
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&user); err != nil {
		logx.GetLogger("study").Errorf("HandlerLogin|Decode err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	err := us.userServer.Login(ctx.Request().Context(), user.Uid, user.Password)
	if err != nil {
		logx.GetLogger("study").Errorf("HandlerLogin|Login Error|%v", err)
		if errors.Is(err, core2.ErrPasswordNotMatch) {
			return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
				"msg": "Password Error",
			})
		} else if errors.Is(err, core2.ErrUserNotExist) {
			return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
				"msg": "User Not Exist",
			})
		} else {
			return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
				"msg": "login Error",
			})
		}
	}

	token, err := us.userServer.CreateToken(ctx.Request().Context(), user.Uid, user.Role)
	if err != nil {
		logx.GetLogger("study").Errorf("HandlerLogin|CreateToken Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "login Error",
		})
	}

	logx.GetLogger("study").Infof("HandlerLogin|Login Success|%s", common.ToStringWithoutError(user))

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, echo.Map{
		"token": token,
		"id":    user.Uid,
		"role":  user.Role,
	})
}

func (us *UserService) HandleSignUp(ctx echo.Context) error {
	var userInfo core2.UserInfo
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&userInfo); err != nil {
		logx.GetLogger("study").Errorf("HandleSignUp|Decode err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}
	// 生成uid
	userInfo.Role = core2.UserRole.Student
	password, err := bcrypt.GenerateFromPassword([]byte(userInfo.Password), bcrypt.DefaultCost)
	if err != nil {
		logx.GetLogger("study").Errorf("HandleSignUp|GenerateFromPassword Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Password Error",
		})
	}
	userInfo.Password = string(password)
	userInfo.CreateTs = time.Now().Unix()
	userInfo.UpdateTs = time.Now().Unix()
	userInfo.Status = core2.UserStatus.Active

	err = us.userServer.RegisterUser(ctx.Request().Context(), &userInfo)
	if err != nil {
		logx.GetLogger("study").Errorf("HandleSignUp|RegisterUser Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "SignUp Error",
		})
	}

	logx.GetLogger("study").Infof("HandleSignUp|Create User Success|%s", common.ToStringWithoutError(userInfo))
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, userInfo)
}

// 发送邮件代码
//func (us *UserService) SendMessage(ctx echo.Context) error {
//	to := ctx.Param("email")
//	subject := ctx.Param("subject")
//
//	htmlTemplate := `<!DOCTYPE html>
//<html lang="zh">
//<head>
//    <meta charset="UTF-8">
//    <meta name="viewport" content="width=device-width, initial-scale=1.0">
//    <title>LearnX</title>
//    <style>
//        body { font-family: Arial, sans-serif; background-color: #f4f4f4; text-align: center; padding: 20px; }
//        .email-container { background: white; padding: 20px; border-radius: 10px; box-shadow: 0 0 10px rgba(0, 0, 0, 0.1); display: inline-block; max-width: 400px; width: 100%; }
//        .verification-code { font-size: 24px; font-weight: bold; color: #007bff; margin: 20px 0; }
//        .footer { margin-top: 15px; font-size: 14px; color: #666; }
//    </style>
//</head>
//<body>
//    <div class="email-container">
//        <h2>您的验证码</h2>
//        <p>请使用以下验证码完成验证：</p>
//        <p class="verification-code">%s</p>
//		<p class="footer">验证码5分钟之内有效，请及时完成验证。</p>
//        <p class="footer">如果您没有请求此验证码，请忽略此邮件。</p>
//    </div>
//</body>
//</html>`
//
//	verfiyCode := GenerateRandomString(6)
//
//	htmlTemplate = fmt.Sprintf(htmlTemplate, verfiyCode)
//
//	message := gomail.NewMessage()
//	message.SetHeader("From", message.FormatAddress(*config.GloableConfig.Email.Sender, *config.GloableConfig.Email.Alias))
//	message.SetHeader("To", to)
//	message.SetHeader("Subject", subject)
//	message.SetBody("text/html", htmlTemplate)
//
//	if err := us.email.DialAndSend(message); err != nil {
//		logx.GetLogger("study").Errorf("HandlerListUsers|SendMessage|Error|%v", err)
//		//return ctx.JSON(http.StatusBadRequest,
//		//	httpx.HttpResponse{StatusCode: httpx.HttpStatusCode.HttpInternalError, Msg: "Send Message Error"})
//		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
//			"msg": "Send Message Error",
//		})
//	}
//
//	// 把生成的验证码存入到redis中
//	err := us.redis.SetNX(ctx.Request().Context(), fmt.Sprintf(core.RedisVerifyCodeKey, to), verfiyCode, time.Minute*5).Err()
//	if nil != err {
//		logx.GetLogger("study").Errorf("HandlerListUsers|VerifyCode Set To Redis|Error|%v", err)
//		//return ctx.JSON(http.StatusBadRequest,
//		//	httpx.HttpResponse{StatusCode: httpx.HttpStatusCode.HttpInternalError, Msg: "Send Message Error"})
//		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
//			"msg": "Send Message Error",
//		})
//	}
//
//	//return ctx.JSON(http.StatusOK, httpx.HttpResponse{StatusCode: httpx.HttpStatusCode.HttpOK, Msg: "Send Message Success"})
//	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, echo.Map{
//		"msg": "Send Message Success",
//	})
//}

func (us *UserService) UpdateUserInfo(ctx echo.Context) error {
	var userInfo core2.UserInfo
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&userInfo); err != nil {
		logx.GetLogger("study").Errorf("HandlerListUsers|ctx.Bind err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}
	err := us.userServer.UpdateUserInfo(ctx.Request().Context(), &userInfo)
	if err != nil {
		logx.GetLogger("study").Errorf("HandlerListUsers|UpdateUserInfo|Update User Info Error|%v", err)
		return err
	}
	logx.GetLogger("study").Errorf("HandlerListUsers|UpdateUserInfo|Update User Info Success|%s", common.ToStringWithoutError(userInfo))
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, echo.Map{
		"msg": "Update User Info Success",
	})
}

func (us *UserService) HandlerQueryUserInfo(ctx echo.Context) error {
	uid := ctx.Get("uid").(string)
	if len(uid) <= 0 {
		logx.GetLogger("study").Errorf("HandlerListUsers|uid is empty")
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	// 数据库中查询用户信息
	info, err := us.userServer.QueryUserInfo(ctx.Request().Context(), uid)
	if err != nil {
		logx.GetLogger("study").Errorf("HandlerListUsers|QueryUserInfo|err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Query User Info Error",
		})
	}

	logx.GetLogger("study").Infof("HandlerListUsers|QueryUserInfo|Succ|%s", common.ToStringWithoutError(info))

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, info)
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
