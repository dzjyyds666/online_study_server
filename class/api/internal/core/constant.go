package core

import "fmt"

const (
	RedisClassVideoKey           = "class:%s:video:list"         // zset [classid]  存储课程对应视频的fid，根据视频上传的时间进行排序
	RedisTeacherClassListKey     = "class:teacher:%s:class:list" // zset [uid] 储存教师对应的课程列表信息，根据时间进行排序
	RedisStudentSubscribeListKey = "class:student:%s:subscribe:list"
	RedisClassSubscribeStuList   = "class:class:%s:subscribe:stu:list"
	RedisClassListKey            = "class:list"
	RedisClassChapterKey         = "class:class:%s:chapter:list"           // zet [classid] 存储课程对应的章节列表
	RedisChapterIndexKey         = "class:class:chapter:%s:index"          // 章节的index信息
	RedisChapterResourceKey      = "class:class:chapter:resource:%s:index" // 章节资源的index文件
)

func BuildClassVideoListKey(classId string) string {
	return fmt.Sprintf(RedisClassVideoKey, classId)
}

func BuildTeacherClassListKey(uid string) string {
	return fmt.Sprintf(RedisTeacherClassListKey, uid)
}

func BuildStudentSubscribeListKey(uid string) string {
	return fmt.Sprintf(RedisStudentSubscribeListKey, uid)
}

func BuildClassSubscribeStuList(classId string) string {
	return fmt.Sprintf(RedisClassSubscribeStuList, classId)
}

func BuildClassChapterKey(classId string) string {
	return fmt.Sprintf(RedisClassChapterKey, classId)
}

func BuildChapterIndexKey(chapterId string) string {
	return fmt.Sprintf(RedisChapterIndexKey, chapterId)
}

func BuildChapterResourceKey(fid string) string {
	return fmt.Sprintf(RedisChapterResourceKey, fid)
}
