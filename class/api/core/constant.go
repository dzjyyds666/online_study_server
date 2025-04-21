package core

import "fmt"

var (
	ErrTaskHasSubmit = fmt.Errorf("Task Has Submit")
)

const (
	RedisAllClassList     = "class:lists" // 所有课程列表
	RedisDeletedClassList = "class:lists:deleted"

	RedisTeacherClassListKey   = "class:teacher:%s:class:list" // 教师对应课程列表
	RedisTeacherClassDeleteKey = "class:teacher:%s:class:deleted:list"

	RedisClassInfoKey       = "class:%s:info"             // 课程信息key
	RedisSourceChaptersKey  = "class:%s:chapters:list"    // 课程对应章节列表
	RedisClassStudyClassKey = "class:%s:study_class:list" // 课程对应教学班列表

	RedisClassStudentsKey = "class:%s:student:list" //课程对应学生列表

	RedisChapterInfoKey      = "class:chapter:%s:info"          // 课程对应章节信息
	RedisChapterResourceList = "class:chapter:%s:resource:list" // 课程对应章节资源列表

	RedisResourceInfoKey = "class:resource:%s:info" // 课程资源信息

	RedisMd5FileList = "class:md5:%s:file:list" // md5文件列表

	RedisClassTaskList = "class:%s:task:list" // 任务列表
	RedisTaskInfo      = "class:task:%s:info" // 任务信息

	RedisStudentTaskInfoKey     = "class:student:task:%s:info"      // 学生提交的作业信息
	RedisTaskStudentListKey     = "class:task:%s:student:list"      // 作业下面提交的学生列表
	RedisTaskStudentTaskListKey = "class:task:%s:student:task:list" // 作业下面提交的学生列表
	RedisStudentTaskListKey     = "class:student:%s:task:list"
)

func BuildClassTaskList(cid string) string {
	return fmt.Sprintf(RedisClassTaskList, cid)
}

func BuildTaskInfo(tid string) string {
	return fmt.Sprintf(RedisTaskInfo, tid)
}

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

func BuildClassStudyClassList(cid string) string {
	return fmt.Sprintf(RedisClassStudyClassKey, cid)
}

func BuildClassDeletedList() string {
	return RedisDeletedClassList
}

func BuildClassStudentList(cid string) string {
	return fmt.Sprintf(RedisClassStudentsKey, cid)
}

func BuildStudentTaskInfoKey(id string) string {
	return fmt.Sprintf(RedisStudentTaskInfoKey, id)
}

func BuildTaskStudentTaskListKey(id string) string {
	return fmt.Sprintf(RedisTaskStudentTaskListKey, id)
}

func BuildTaskStudentListKey(id string) string {
	return fmt.Sprintf(RedisTaskStudentListKey, id)
}

func BuildStudentTaskListKey(uid string) string {
	return fmt.Sprintf(RedisStudentTaskListKey, uid)
}
