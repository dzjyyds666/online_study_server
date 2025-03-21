package core

import "fmt"

//const (
//	RedisClassVideoKey           = "class:%s:video:list"         // zset [classid]  存储课程对应视频的fid，根据视频上传的时间进行排序
//	RedisTeacherClassListKey     = "class:teacher:%s:class:list" // zset [uid] 储存教师对应的课程列表信息，根据时间进行排序
//	RedisTeaherClassDeletedKey   = "class:teacher:%s:class:deleted:list"
//	RedisStudentSubscribeListKey = "class:student:%s:subscribe:list"
//
//	RedisClassSubscribeStuList = "class:class:%s:subscribe:stu:list"
//	RedisClassListKey          = "class:list"
//	RedisClassChapterKey       = "class:class:%s:chapter:list" // zet [classid] 存储课程对应的章节列表
//	RedisClassInfoKey          = "class:class:%s:info"
//
//	RedisChapterIndexKey         = "class:class:chapter:%s:index"          // 章节的index信息
//	RedisChapterResourceIndexKey = "class:class:chapter:resource:%s:index" // 章节资源的index文件
//	RedisChapterResourceKey      = "class:class:chapter:%s:resource:list"  // zset [chid] 章节的资源列表
//
//	RedisStudyClassTeacherKey = "class:%s:study_class:list" // 老师对应教学班列表 set
//	REdisStudyClassClassKey   = "class:%s:study_class:list" // 课程对应的教学班列表 set
//	RedisStudyClassInfoKey    = "class:study_class:%s:info"
//)
//
//func BuildClassVideoListKey(classId string) string {
//	return fmt.Sprintf(RedisClassVideoKey, classId)
//}
//
//func BuildTeacherClassListKey(uid string) string {
//	return fmt.Sprintf(RedisTeacherClassListKey, uid)
//}
//
//func BuildStudentSubscribeListKey(uid string) string {
//	return fmt.Sprintf(RedisStudentSubscribeListKey, uid)
//}
//
//func BuildClassSubscribeStuList(classId string) string {
//	return fmt.Sprintf(RedisClassSubscribeStuList, classId)
//}
//
//func BuildClassChapterKey(classId string) string {
//	return fmt.Sprintf(RedisClassChapterKey, classId)
//}
//
//func BuildChapterIndexKey(chapterId string) string {
//	return fmt.Sprintf(RedisChapterIndexKey, chapterId)
//}
//
//func BuildChapterResourceKey(chapterId string) string {
//	return fmt.Sprintf(RedisChapterResourceKey, chapterId)
//}
//
//func BuildChapterResourceIndexKey(fid string) string {
//	return fmt.Sprintf(RedisChapterResourceIndexKey, fid)
//}
//
//func BuildClassInfoKey(classId string) string {
//	return fmt.Sprintf(RedisClassInfoKey, classId)
//}
//
//func BuildStudyClassTeacherKey(tid string) string {
//	return fmt.Sprintf(RedisStudyClassTeacherKey, tid)
//}
//
//func BuildStudyClassClassKey(cid string) string {
//	return fmt.Sprintf(REdisStudyClassClassKey, cid)
//}
//
//func BuildStudyClassInfoKey(scid string) string {
//	return fmt.Sprintf(RedisStudyClassInfoKey, scid)
//}
//
//func BuildTeahcerClassDeletedKey(tid string) string {
//	return fmt.Sprintf(RedisTeaherClassDeletedKey, tid)
//}

const (
	RedisAllClassList = "class:lists" // 所有课程列表

	RedisTeacherClassListKey   = "class:teacher:%s:class:list" // 教师对应课程列表
	RedisTeacherClassDeleteKey = "class:teacher:%s:class:deleted:list"

	RedisClassInfoKey       = "class:%s:info"             // 课程信息key
	RedsiClassChaptersKey   = "class:%s:chapters:list"    // 课程对应章节列表
	RedisClassStudyClassKey = "class:%s:study_class:list" // 课程对应教学班列表

	RedisStudyClassInfoKey = "class:study_class:%s:info" // 教学班信息key

	RedisChapterInfoKey      = "class:chapter:%s:info"          // 课程对应章节信息
	RedisChapterResourceList = "class:chapter:%s:resource:list" // 课程对应章节资源列表

	RedisResourceInfoKey = "class:resource:%s:info" // 课程资源信息

	RedisMd5FileList = "class:md5:%s:file:list" // md5文件列表
)

func BuildAllClassList() string {
	return RedisAllClassList
}

func BuildTeacherClassList(uid string) string {
	return fmt.Sprintf(RedisTeacherClassListKey, uid)
}

func BuildTeacherClassDeletedList(uid string) string {
	return fmt.Sprintf(RedisTeacherClassDeleteKey, uid)
}

func BuildClassInfo(cid string) string {
	return fmt.Sprintf(RedisClassInfoKey, cid)
}

func BuildClassChapterList(cid string) string {
	return fmt.Sprintf(RedsiClassChaptersKey, cid)
}

func BuildChapterInfo(chid string) string {
	return fmt.Sprintf(RedisChapterInfoKey, chid)
}

func BuildChapterResourceList(chid string) string {
	return fmt.Sprintf(RedisChapterResourceList, chid)
}

func BuildResourceInfo(fid string) string {
	return fmt.Sprintf(RedisResourceInfoKey, fid)
}

func BuildMd5FileList(md5 string) string {
	return fmt.Sprintf(RedisMd5FileList, md5)
}

func BuildStudyClassInfo(scid string) string {
	return fmt.Sprintf(RedisStudyClassInfoKey, scid)
}

func BuildClassStudyClassList(cid string) string {
	return fmt.Sprintf(RedisClassStudyClassKey, cid)
}
