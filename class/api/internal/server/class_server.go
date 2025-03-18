package server

import (
	"class/api/config"
	"class/api/internal/core"
	"crypto/rand"
	"encoding/json"
	"errors"
	"github.com/dzjyyds666/opensource/httpx"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/labstack/echo"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"strconv"
	"time"
)

type ClassServer struct {
	redis *redis.Client
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

	//// 连接mysql
	//dsn := *config.GloableConfig.Mysql.Username + ":" + *config.GloableConfig.Mysql.Password + "@tcp(" + *config.GloableConfig.Mysql.Host + ":" + strconv.Itoa(*config.GloableConfig.Mysql.Port) + ")/" + *config.GloableConfig.Mysql.Database + "?charset=utf8mb4&parseTime=True&loc=Local"
	//msClient, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	//if err != nil {
	//	logx.GetLogger("OS_Server").Errorf("NewClassServer|gorm.Open err:%v", err)
	//	return nil, err
	//}

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
	var class core.Class
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&class); err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerCreateClass|ctx.Bind err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Param Invalid",
		})
	}

	tuid := ctx.Get("uid")

	//生成随机的classId
	cid := "ci_" + GenerateRandomString(8)
	class.WithCid(cid).
		WithCreateTs(time.Now().Unix()).
		WithDeleted(false).
		WithArchive(false).
		WithTeacher(tuid.(string))

	marshal, err := json.Marshal(&class)
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerCreateClass|Marshal Class Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "system internal",
		})
	}

	err = cls.redis.Set(ctx.Request().Context(), core.BuildClassInfoKey(cid), marshal, 0).Err()
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerCreateClass|Marshal Class Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "create class error",
		})
	}

	// 把该课程加入教师的课程列表中
	err = cls.redis.ZAdd(ctx.Request().Context(), core.BuildTeacherClassListKey(tuid.(string)), redis.Z{
		Score:  float64(*class.CreateTs),
		Member: cid,
	}).Err()

	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerCreateClass|Add Class To Teacher List Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "insert Redis Error",
		})
	}

	// 把该课程加入到课程list中
	_, err = cls.redis.ZAdd(ctx.Request().Context(), core.RedisClassListKey, redis.Z{
		Score:  float64(*class.CreateTs),
		Member: cid,
	}).Result()
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerCreateClass|Add Class To ClassList Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "insert Redis Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, class)
}

func (cls *ClassServer) HandleListClass(ctx echo.Context) error {
	var classLists core.ClassList
	if err := ctx.Bind(&classLists); err != nil {
		logx.GetLogger("OS_Server").Errorf("HandleListClass|ctx.Bind err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Param Invalid",
		})
	}
	uid := ctx.Get("uid")
	classLists, err := classLists.QueryClassList(ctx.Request().Context(), uid.(string), cls.redis)
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandleListClass|QueryClassList Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Query Class List Error",
		})
	}
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, classLists)
}

func (cls *ClassServer) HandlerUpdateClass(ctx echo.Context) error {
	var class core.Class
	err := ctx.Bind(&class)
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerUpdateClass|ctx.Bind err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Param Invalid",
		})
	}

	if class.Cid == nil {
		logx.GetLogger("OS_Server").Errorf("HandlerUpdateClass|ctx.Bind err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "classid can not be null",
		})
	}

	marshal, err := json.Marshal(&class)
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerUpdateClass|Marshal Class Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "system internal",
		})
	}

	err = cls.redis.Set(ctx.Request().Context(), core.BuildClassInfoKey(*class.Cid), marshal, 0).Err()
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerUpdateClass|Marshal Class Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "system internal",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, class)
}

// 上传完课程信息之后
// 把课程移入垃圾箱
func (cls *ClassServer) HandlerPutClassInTrash(ctx echo.Context) error {
	var class core.Class
	err := ctx.Bind(&class)
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerDeleteClass|ctx.Bind err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Param Invalid",
		})
	}

	//// 先查询这个课程下面是否有没有删除的视频，如果有，不允许删除
	//videoKey := core.BuildClassVideoListKey(*class.Cid)
	//videoCount, err := cls.redis.ZCard(ctx.Request().Context(), videoKey).Result()
	//if err != nil {
	//	logx.GetLogger("OS_Server").Errorf("HandlerDeleteClass|Get Video Count Error|%v", err)
	//	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
	//		"msg": "Get Video Count Error",
	//	})
	//}
	//
	//if videoCount > 0 {
	//	logx.GetLogger("OS_Server").Errorf("HandlerDeleteClass|Has Video|%v", err)
	//	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpBadRequest, echo.Map{
	//		"msg": "Has Video",
	//	})
	//}

	// 更新mysql的数据，把课程的删除为修改为delete
	result, err := cls.redis.Get(ctx.Request().Context(), core.BuildClassInfoKey(*class.Cid)).Result()
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerDeleteClass|Get Class Info Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Get Class Info Error",
		})
	}
	err = json.Unmarshal([]byte(result), &class)
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerDeleteClass|Unmarshal Class Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Unmarshal Class Error",
		})
	}

	class.WithDeleted(true)
	marshal, _ := json.Marshal(&class)
	err = cls.redis.Set(ctx.Request().Context(), core.BuildClassInfoKey(*class.Cid), marshal, 0).Err()
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerDeleteClass|Update Class Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Update Class Error",
		})
	}

	// 从classlist中移除
	err = cls.redis.ZRem(ctx.Request().Context(), core.RedisClassListKey, class.Cid).Err()
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerDeleteClass|Remove Class From ClassList Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Remove Class From ClassList Error",
		})
	}

	// 从老师的正常课程列表清除
	teacherKey := core.BuildStudyClassTeacherKey(*class.Teacher)
	err = cls.redis.ZRem(ctx.Request().Context(), teacherKey, class.Cid).Err()
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerDeleteClass|Remove Class From Teacher List Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Remove Class From Teacher List Error",
		})
	}

	// 从加入老师的删除列表中
	teacherKey = core.BuildTeahcerClassDeletedKey(*class.Teacher)
	err = cls.redis.ZAdd(ctx.Request().Context(), teacherKey, redis.Z{
		Member: class.Cid,
		Score:  float64(time.Now().Unix()),
	}).Err()
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerDeleteClass|Add Class To Teacher Deleted List Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Add Class To Teacher Deleted List Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, class)
}

// 恢复课程
func (cls *ClassServer) HandlerRecoverClass(ctx echo.Context) error {
	cid := ctx.Param("cid")
	if len(cid) <= 0 {
		logx.GetLogger("OS_Server").Errorf("HandlerRecoverClass|cid is empty")
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "cid is empty",
		})
	}

	// 修改数据库中的内容
	var class core.Class
	result, err := cls.redis.Get(ctx.Request().Context(), core.BuildClassInfoKey(cid)).Result()
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerRecoverClass|Get Class Info Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Get Class Info Error",
		})
	}

	err = json.Unmarshal([]byte(result), &class)
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerRecoverClass|Unmarshal Class Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Unmarshal Class Error",
		})
	}

	class.WithDeleted(false)

	marshal, _ := json.Marshal(&class)
	err = cls.redis.Set(ctx.Request().Context(), core.BuildClassInfoKey(cid), marshal, 0).Err()
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerRecoverClass|Update Class Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Update Class Error",
		})
	}

	// 把课程添加到classlist中
	err = cls.redis.ZAdd(ctx.Request().Context(), core.RedisClassListKey, redis.Z{
		Score:  float64(*class.CreateTs),
		Member: cid,
	}).Err()
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerRecoverClass|Add Class To ClassList Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Add Class To ClassList Error",
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
	var class core.Class
	result, err := cls.redis.Get(ctx.Request().Context(), core.BuildClassInfoKey(cid)).Result()
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerDeleteClass|Get Class Info Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Get Class Info Error",
		})
	}

	err = json.Unmarshal([]byte(result), &class)
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerDeleteClass|Unmarshal Class Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Unmarshal Class Error",
		})
	}

	uid := ctx.Get("uid")
	if class.Teacher != uid {
		logx.GetLogger("OS_Server").Errorf("HandlerDeleteClass|Not Owner|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpBadRequest, echo.Map{
			"msg": "You Can Not Delete other's class",
		})
	}

	if !*class.Deleted {
		logx.GetLogger("OS_Server").Errorf("HandlerDeleteClass|Class Not In Trash|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpBadRequest, echo.Map{
			"msg": "Class Not In Trash",
		})
	}

	// redis老师的课表中删除改课程
	key := core.BuildTeacherClassListKey(*class.Teacher)
	err = cls.redis.ZRem(ctx.Request().Context(), key, cid).Err()
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerDeleteClass|Delete Class From Teacher List Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Delete Class From Teacher List Error",
		})
	}

	err = cls.redis.Del(ctx.Request().Context(), core.BuildClassInfoKey(cid)).Err()
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerDeleteClass|Delete Class From ClassList Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Delete Class From ClassList Error",
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

	var class core.Class
	classInfo, err := cls.redis.Get(ctx.Request().Context(), core.BuildClassInfoKey(cid)).Result()
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerDeleteClass|Query Class Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Query Class Error",
		})
	}

	err = json.Unmarshal([]byte(classInfo), &class)
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerDeleteClass|Query Class Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Query Class Error",
		})
	}

	// 查看本人的订阅中有没有该class
	uid := ctx.Get("uid")
	key := core.BuildStudentSubscribeListKey(uid.(string))
	result, err := cls.redis.ZScore(ctx.Request().Context(), key, cid).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		logx.GetLogger("OS_Server").Errorf("HandlerDeleteClass|Query Class Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Query Class Error",
		})
	}

	// 查询订阅人数
	classKey := core.BuildClassSubscribeStuList(cid)
	subscribeCount, err := cls.redis.ZCard(ctx.Request().Context(), classKey).Result()
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerDeleteClass|Query Class Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Query Class Error",
		})
	}

	idSubscribe := false

	if result > 0 {
		idSubscribe = true
	}

	// todo rpc调用用户信息查询教师信息

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, echo.Map{
		"class": class,
		//"teacher":teacher,
		"idSubscribe":    idSubscribe,
		"subscribeCount": subscribeCount,
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
	err := cls.redis.ZAdd(ctx.Request().Context(), stuKey, redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: cid,
	}).Err()
	if err != nil {
		logx.GetLogger("OS_Server").Infof("HandlerSubscribeClass|Add Subscribe Class Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Add Subscribe Class Error",
		})
	}

	// 把学生写入课程的订阅学生列表
	classKey := core.BuildClassSubscribeStuList(cid)
	err = cls.redis.ZAdd(ctx.Request().Context(), classKey, redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: uid,
	}).Err()

	if err != nil {
		logx.GetLogger("OS_Server").Infof("HandlerSubscribeClass|Add Subscribe Class Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Add Subscribe Class Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, echo.Map{
		"msg": "Subscribe Class Success",
	})
}

func (cls *ClassServer) HandlerCancleSubscribeClass(ctx echo.Context) error {
	cid := ctx.Param("cid")
	if len(cid) <= 0 {
		logx.GetLogger("OS_Server").Errorf("HandlerDeleteClass|cid is empty")
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "cid is empty",
		})
	}
	uid := ctx.Get("uid")

	// 去除学生的订阅课程列表
	stuKey := core.BuildStudentSubscribeListKey(uid.(string))
	err := cls.redis.ZRem(ctx.Request().Context(), stuKey, cid).Err()
	if err != nil {
		logx.GetLogger("OS_Server").Infof("HandlerSubscribeClass|Cancle Subscribe Class Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Cancle Subscribe Class Error",
		})
	}

	// 去除课程的订阅学生列表
	classKey := core.BuildClassSubscribeStuList(cid)
	err = cls.redis.ZRem(ctx.Request().Context(), classKey, uid).Err()
	if err != nil {
		logx.GetLogger("OS_Server").Infof("HandlerSubscribeClass|Cancle Subscribe Class Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Cancle Subscribe Class Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, echo.Map{
		"msg": "Cancle Subscribe Class Success",
	})
}

func (cls *ClassServer) HandlerCreateChapter(ctx echo.Context) error {
	var chapter core.Chapter
	cid := ctx.Param("cid")
	if err := ctx.Bind(chapter); err != nil || len(cid) <= 0 {
		logx.GetLogger("OS_Server").Errorf("HandlerListUsers|ctx.Bind err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	// 生成章节id
	chid := "ch_" + GenerateRandomString(8)
	chapter.Chid = &chid
	createTs := time.Now().Unix()
	chapter.CreateTs = &createTs

	err := chapter.CreateChapter(ctx.Request().Context(), cid, cls.redis)
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerCreateChapter|Create Chapter Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Create Chapter Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, chapter)
}

func (cls *ClassServer) HandlerRenameChapter(ctx echo.Context) error {
	var chapter core.Chapter
	if err := ctx.Bind(chapter); err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerRenameChapter|ctx.Bind err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	err := chapter.RanameChapter(ctx.Request().Context(), *chapter.ChapterName, cls.redis)
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerRenameChapter|Rename Chapter Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Rename Chapter Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, chapter)
}

func (cls *ClassServer) HandlerUploadResuorce(ctx echo.Context) error {
	var resource core.ClassResource
	chid := ctx.Param("chid")
	if err := ctx.Bind(&resource); err != nil || len(chid) <= 0 {
		logx.GetLogger("OS_Server").Errorf("HandlerUploadResuorce|ctx.Bind err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	err := resource.CreateResource(ctx.Request().Context(), chid, cls.redis)
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandlerUploadResuorce|Create Resource Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "upload Resource Error",
		})
	}
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, resource)
}

func (cls *ClassServer) HandlerDeleteResource(ctx echo.Context) error {
	fid := ctx.Param("fid")
	if len(fid) <= 0 {
		logx.GetLogger("OS_Server").Errorf("HandlerDeleteResource|fid is empty")
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "fid is empty",
		})
	}

	var resource core.ClassResource
	resource.Fid = &fid

	resource.DeleteResource(ctx.Request().Context(), cls.redis)

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, echo.Map{
		"msg": "delete resource success",
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

func (cls *ClassServer) HandleCreateStudyClass(ctx echo.Context) error {
	var studyClass core.StudyClass
	if err := ctx.Bind(&studyClass); err != nil {
		logx.GetLogger("OS_Server").Errorf("HandleCreateStudyClass|ctx.Bind err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	studyClass.SCid = "sc_" + GenerateRandomString(8)

	err := studyClass.CreateStudyClass(ctx.Request().Context(), cls.redis)
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandleCreateStudyClass|Create StudyClass Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Create StudyClass Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, studyClass)
}

func (cls *ClassServer) HandleQueryStudyClass(ctx echo.Context) error {
	var studyClassList core.StudyClassList
	if err := ctx.Bind(&studyClassList); err != nil {
		logx.GetLogger("OS_Server").Errorf("HandleQueryStudyClass|ctx.Bind err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	list, err := studyClassList.QueryStudyClassList(ctx.Request().Context(), cls.redis)
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandleQueryStudyClass|Query StudyClass Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Query StudyClass Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, list)
}

func (cls *ClassServer) HandleDeleteStudyClass(ctx echo.Context) error {
	var studyClass core.StudyClass
	if err := ctx.Bind(&studyClass); err != nil {
		logx.GetLogger("OS_Server").Errorf("HandleDeleteStudyClass|ctx.Bind err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	err := studyClass.DeleteStudyClass(ctx.Request().Context(), cls.redis)
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("HandleDeleteStudyClass|Delete StudyClass Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Delete StudyClass Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, studyClass)
}
