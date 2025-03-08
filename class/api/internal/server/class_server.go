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
	class.CreateTs = time.Now().Unix()
	class.UpdateTs = time.Now().Unix()

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

type classVideo struct {
	Cid string `json:"cid"`
	Fid string `json:"fid"`
}

func (cls *ClassServer) HandlerClassAddVideo(ctx echo.Context) error {
	// 在reids中插入课程的视频
	var upload classVideo
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

	var class Class
	class.Cid = upload.Cid
	class.UpdateTs = time.Now().Unix()

	// 修改mysql的更新时间
	err = cls.mysql.Model(&class).Where("cid = ?", upload.Cid).Updates(&class).Error
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerUpdateClass|Update Class Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Update Class Error",
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

	class.UpdateTs = time.Now().Unix()

	err = cls.mysql.Model(&class).Where("cid = ?", class.Cid).Updates(&class).Error
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerUpdateClass|Update Class Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Update Class Error",
		})
	}
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, class)
}

// 删除课程的某一节视频 传过来视频的fid class对应的视频列表下面删除掉视频，然后rpc调用存储服务删除文件
func (cls *ClassServer) HandlerDeleteVideo(ctx echo.Context) error {

	var video classVideo
	err := ctx.Bind(&video)
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerDeleteVideo|ctx.Bind err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Param Invalid",
		})
	}

	uid := ctx.Get("uid")

	// 先校验一下改课程是不是这个老师的课程，如果不是，返回错误
	var class Class
	err = cls.mysql.Where("cid = ?", video.Cid).First(&class).Error
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerDeleteVideo|QueryClassInfoError|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Query Classinfo Error",
		})
	}

	if class.Owner != uid {
		logx.GetLogger("OS_Server").Errorf("HandlerDeleteVideo|NotOwner|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpBadRequest, echo.Map{
			"msg": "You Can Not Edit other's class",
		})
	}

	key := core.BuildClassVideoListKey(video.Cid)

	err = cls.redis.ZRem(ctx.Request().Context(), key, video.Fid).Err()
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerDeleteVideo|DeleteVideoError|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Delete Video Error",
		})
	}

	class.UpdateTs = time.Now().Unix()
	err = cls.mysql.Model(&class).Where("cid = ?", video.Cid).Updates(&class).Error
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerDeleteVideo|Update Class Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Update Class Error",
		})
	}

	// todo 调用rpc代码删除文件

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, video)
}

// 上传完课程信息之后
// 把课程移入垃圾箱
func (cls *ClassServer) HandlerPutClassInTrash(ctx echo.Context) error {
	var class Class
	err := ctx.Bind(&class)
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerDeleteClass|ctx.Bind err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Param Invalid",
		})
	}

	// 先查询这个课程下面是否有没有删除的视频，如果有，不允许删除
	videoKey := core.BuildClassVideoListKey(class.Cid)
	videoCount, err := cls.redis.ZCard(ctx.Request().Context(), videoKey).Result()
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerDeleteClass|Get Video Count Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Get Video Count Error",
		})
	}

	if videoCount > 0 {
		logx.GetLogger("OS_Server").Errorf("HandlerDeleteClass|Has Video|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpBadRequest, echo.Map{
			"msg": "Has Video",
		})
	}

	uid := ctx.Get("uid")

	// 更新mysql的数据，把课程的删除为修改为delete
	result := cls.mysql.
		Model(&class).
		Where("cid = ? AND owner = ?", class.Cid, uid).
		Updates(map[string]interface{}{"is_delete": ClassStatus.Delete, "update_ts": time.Now().Unix()})

	if result.Error != nil {
		logx.GetLogger("OS_Server").Infof("HandlerDeleteClass|Update Class Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Update Class Error",
		})
	}

	if result.RowsAffected == 0 {
		logx.GetLogger("OS_Server").Errorf("HandlerDeleteClass|Update Class Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpBadRequest, echo.Map{
			"msg": "no class match",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, class)
}

func (cls *ClassServer) HandlerDeleteClass(ctx echo.Context) error {
	cid := ctx.Param("cid")
	if len(cid) <= 0 {
		logx.GetLogger("OS_Server").Errorf("HandlerDeleteClass|cid is empty")
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "cid is empty",
		})
	}

	// mysql中查询数据
	var class Class
	err := cls.mysql.Where("cid = ?", cid).First(&class).Error
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerDeleteClass|Query Class Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Query Class Error",
		})
	}

	uid := ctx.Get("uid")
	if class.Owner != uid {
		logx.GetLogger("OS_Server").Errorf("HandlerDeleteClass|Not Owner|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpBadRequest, echo.Map{
			"msg": "You Can Not Delete other's class",
		})
	}

	if class.IsDelete != ClassStatus.Delete {
		logx.GetLogger("OS_Server").Errorf("HandlerDeleteClass|Class Not In Trash|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpBadRequest, echo.Map{
			"msg": "Class Not In Trash",
		})
	}

	// redis老师的课表中删除改课程
	key := core.BuildTeacherClassListKey(class.Owner)
	err = cls.redis.ZRem(ctx.Request().Context(), key, cid).Err()
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerDeleteClass|Delete Class From Teacher List Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Delete Class From Teacher List Error",
		})
	}

	// mysql中删除课程
	err = cls.mysql.Where("cid = ?", cid).Delete(&class).Error
	if err != nil {
		logx.GetLogger("OS_Server").Infof("HandlerDeleteClass|Delete Class Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Delete Class Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, echo.Map{
		"msg": "Delete Class Success",
		"cid": cid,
	})
}

func (cls *ClassServer) HandlerQueryClassInfo(ctx echo.Context) error {
	cid := ctx.Param("cid")
	if len(cid) <= 0 {
		logx.GetLogger("OS_Server").Errorf("HandlerDeleteClass|cid is empty")
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "cid is empty",
		})
	}

	var class Class
	err := cls.mysql.Where("cid = ?", cid).First(&class).Error
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerDeleteClass|Query Class Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Query Class Error",
		})
	}

	if class.IsDelete == ClassStatus.Delete {
		logx.GetLogger("OS_Server").Errorf("HandlerDeleteClass|Class is deleted|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpBadRequest, echo.Map{
			"msg": "Class is deleted",
		})
	}

	// todo rpc调用用户信息查询教师信息

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, echo.Map{
		"class": class,
		//"teacher":teacher,
	})
}

func (cls *ClassServer) HandlerSubscribeClass(ctx echo.Context) error {
	cid := ctx.Param("cid")
	if len(cid) <= 0 {
		logx.GetLogger("OS_Server").Errorf("HandlerDeleteClass|cid is empty")
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "cid is empty",
		})
	}

	uid := ctx.Get("uid")

	// 写入学生的订阅课程列表
	stuKey := core.BuildStudentSubscribeListKey(uid.(string))
	err := cls.redis.SAdd(ctx.Request().Context(), stuKey, cid).Err()
	//if err
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
