package core

import "fmt"

const (
	RedisAllClassList     = "class:lists" // 所有课程列表
	RedisDeletedClassList = "class:lists:deleted"

	RedisTeacherClassListKey   = "class:teacher:%s:class:list" // 教师对应课程列表
	RedisTeacherClassDeleteKey = "class:teacher:%s:class:deleted:list"

	RedisClassInfoKey       = "class:%s:info"             // 课程信息key
	RedisSourceChaptersKey  = "class:%s:chapters:list"    // 课程对应章节列表
	RedisClassStudyClassKey = "class:%s:study_class:list" // 课程对应教学班列表

	RedisStudyClassInfoKey        = "class:study_class:%s:info" // 教学班信息key
	RedisStudyClassStudentListKey = "class:study_class:%s:student:list"

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
	return fmt.Sprintf(RedisSourceChaptersKey, cid)
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

func BuildClassDeletedList() string {
	return RedisDeletedClassList
}

func BUildStudyClassStudentList(scid string) string {
	return fmt.Sprintf(RedisStudyClassStudentListKey, scid)
}
