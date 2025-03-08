package server

import (
	"class/api/config"
	"class/api/internal/core"
	"crypto/rand"
	"github.com/dzjyyds666/opensource/httpx"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/labstack/echo"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strconv"
	"time"
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
	//
	//mongodsn := "mongodb://" + *config.GloableConfig.Mongo.Username + ":" + *config.GloableConfig.Mongo.Password + "@ip:" + *config.GloableConfig.Mongo.Host + "/" + strconv.Itoa(*config.GloableConfig.Mongo.Port)
	//// 连接mongo
	//mgClient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongodsn))
	//if err != nil {
	//	logx.GetLogger("OS_Server").Errorf("NewClassServer|mongo.Connect err:%v", err)
	//	return nil, err
	//}
	//database := mgClient.Database(*config.GloableConfig.Mongo.DB)

	return &ClassServer{
		redis: dsClient,
		mysql: msClient,
		//mongo: database,
	}, nil
}

/*
只需要传递3个参数

	    ClassName  string `json:"class_name" gorm:"class_name"`
		ClassType  string `json:"class_type" gorm:"class_type"`
		ClassDesc  string `json:"class_desc" gorm:"class_desc"`
*/
func (cls *ClassServer) HandlerCreateClass(ctx echo.Context) error {
	var class Class
	err := ctx.Bind(&class)
	if err == nil {
		logx.GetLogger("OS_Server").Errorf("HandlerCreateClass|ctx.Bind err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Param Invalid",
		})
	}

	tuid := ctx.Get("owner").(string)

	//生成随机的classId
	cid := "cl_" + GenerateRandomString(8)
	class.Cid = cid
	class.IsComplete = ClassStatus.UnComplete
	class.IsDelete = ClassStatus.NotDelete
	class.Owner = tuid

	// 插入mysql
	err = cls.mysql.Create(&class).Error
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerCreateClass|Create Class Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "insert Mysql Error",
		})
	}

	// 把该课程加入教师的课程列表中
	err = cls.redis.ZAdd(ctx.Request().Context(), core.BuildTeacherClassListKey(tuid), redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: cid,
	}).Err()

	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerCreateClass|Add Class To Teacher List Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "insert Redis Error",
		})
	}
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, class)
}

type uploadVideo struct {
	Cid string `json:"cid"`
	Fid string `json:"fid"`
}

func (cls *ClassServer) HandlerUploadClass(ctx echo.Context) error {
	// 在reids中插入课程的视频
	var upload uploadVideo
	err := ctx.Bind(&upload)
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("UploadClass|ctx.Bind err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Param Invalid",
		})
	}
	key := core.BuildClassVideoListKey(upload.Cid)
	err = cls.redis.ZAdd(ctx.Request().Context(), key, redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: upload.Fid,
	}).Err()

	if err != nil {
		logx.GetLogger("OS_Server").Errorf("UploadClass|Add Class To Teacher List Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "upload class video err",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, upload)
}

func (cls *ClassServer) HandlerUpdateClass(ctx echo.Context) error {
	var class Class
	err := ctx.Bind(&class)
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerUpdateClass|ctx.Bind err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Param Invalid",
		})
	}

	if class.Cid == "" {
		logx.GetLogger("OS_Server").Errorf("HandlerUpdateClass|ctx.Bind err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "classid can not be null",
		})
	}

	err = cls.mysql.Model(&class).Where("cid = ?", class.Cid).Updates(&class).Error
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerUpdateClass|Update Class Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Update Class Error",
		})
	}
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, class)
}

// 删除课程的某一节视频
func (cls *ClassServer) HandlerDeleteVideo(ctx echo.Context) error {
	return nil
}

// 删除这个课程
func (cls *ClassServer) HandlerDeleteClass(ctx echo.Context) error {
	return nil
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
