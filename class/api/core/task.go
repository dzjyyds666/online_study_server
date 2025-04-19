package core

import "encoding/json"

type Task struct {
	TaskId        string   `json:"task_id"`
	TaskName      string   `json:"task_name"`
	TaskContent   string   `json:"task_content"`
	TaskImageList []string `json:"task_image_list"` // 存储任务的图片列表
	Cid           string   `json:"cid"`
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

func (t *Task) WithImageList(image string) *Task {
	if t.TaskImageList == nil {
		t.TaskImageList = make([]string, 0)
	}
	t.TaskImageList = append(t.TaskImageList, image)
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

type SubmitTask struct {
	Id            string   `json:"id"`
	Content       string   `json:"content"`
	TaskId        string   `json:"task_id"`
	TaskImageList []string `json:"task_image_list"` // 存储任务的图片列表
	Owner         string   `json:"owner"`
	OwnerName     string   `json:"owner_name"`
	Viewing       bool     `json:"viewing"`  // 老师是否查看
	Annotate      string   `json:"annotate"` // 老师的批注
	Level         string   `json:"level"`    // 提交任务的等级
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

func (st *SubmitTask) WithImageList(image string) *SubmitTask {
	if st.TaskImageList == nil {
		st.TaskImageList = make([]string, 0)
	}
	st.TaskImageList = append(st.TaskImageList, image)
	return st
}

func (st *SubmitTask) Marshal() (string, error) {
	marshal, err := json.Marshal(st)
	return string(marshal), err
}

type ListStudentList struct {
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
