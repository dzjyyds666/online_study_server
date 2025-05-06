package core

import "encoding/json"

// 老师提交的作业信息
type Task struct {
	TaskId         string   `json:"task_id" bson:"_id"`
	TaskName       string   `json:"task_name" bson:"task_name"`
	TaskContent    string   `json:"task_content" bson:"task_content"`
	AttachemntList []string `json:"attachment_list" bson:"attachment_list,omitempty"` // 存储任务的图片列表
	Cid            string   `json:"cid" bson:"cid"`
}

func (t *Task) WithId(id string) *Task {
	t.TaskId = id
	return t
}

func (t *Task) WithCid(cid string) *Task {
	t.Cid = cid
	return t
}

func (t *Task) WithName(name string) *Task {
	t.TaskName = name
	return t
}

func (t *Task) WithContent(content string) *Task {
	t.TaskContent = content
	return t
}

func (t *Task) WithAttchmentList(attachment string) *Task {
	if t.AttachemntList == nil {
		t.AttachemntList = make([]string, 0)
	}
	t.AttachemntList = append(t.AttachemntList, attachment)
	return t
}

func (t *Task) Marshal() string {
	marshal, _ := json.Marshal(t)
	return string(marshal)
}

type ListTask struct {
	Tasks   []*Task `json:"tasks"`
	ReferId string  `json:"refer_id"`
	Limit   int64   `json:"limit"`
	Cid     string  `json:"cid"`
}

// 学生提交的作业信息，包括一些状态
type SubmitTask struct {
	Id             string   `json:"id" bson:"_id"`
	Content        string   `json:"content" bson:"content,omitempty"`
	TaskId         string   `json:"task_id" bson:"task_id,omitempty"`
	AttachmentList []string `json:"attachment_list,omitempty" bson:"attachment-list,omitempty"` // 存储任务的图片列表
	Owner          string   `json:"owner" bson:"owner,omitempty"`
	OwnerName      string   `json:"owner_name,omitempty" bson:"owner_name,omitempty"`
	Viewing        bool     `json:"viewing" bson:"viewing,omitempty"`   // 老师是否查看
	Annotate       string   `json:"annotate" bson:"annotate,omitempty"` // 老师的批注
	Level          string   `json:"level" bson:"level,omitempty"`       // 提交任务的等级
}

func (st *SubmitTask) WithId(id string) *SubmitTask {
	st.Id = id
	return st
}

func (st *SubmitTask) WithViewing(viewing bool) *SubmitTask {
	st.Viewing = viewing
	return st
}

func (st *SubmitTask) WithAnnotate(annotate string) *SubmitTask {
	st.Annotate = annotate
	return st
}
func (st *SubmitTask) WithLevel(level string) *SubmitTask {
	st.Level = level
	return st
}

func (st *SubmitTask) WithContent(content string) *SubmitTask {
	st.Content = content
	return st
}

func (st *SubmitTask) WithTaskId(taskId string) *SubmitTask {
	st.TaskId = taskId
	return st
}

func (st *SubmitTask) WithOwner(owner string) *SubmitTask {
	st.Owner = owner
	return st
}

func (st *SubmitTask) WithAttachmentList(resource string) *SubmitTask {
	if st.AttachmentList == nil {
		st.AttachmentList = make([]string, 0)
	}
	st.AttachmentList = append(st.AttachmentList, resource)
	return st
}

func (st *SubmitTask) Marshal() (string, error) {
	marshal, err := json.Marshal(st)
	return string(marshal), err
}

type ListStudentTask struct {
	Tasks  []*SubmitTask `json:"tasks"`
	Page   int64         `json:"page"`
	Limit  int64         `json:"limit"`
	TaskId string        `json:"task_id"`
}

type ListOwnerTask struct {
	Submit    *SubmitTask `json:"submit"`
	Task      *Task       `json:"task"`
	ClassInfo *Class      `json:"class_info"`
}
