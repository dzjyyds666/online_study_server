package core

import "encoding/json"

type Task struct {
	TaskId      string `json:"task_id"`
	TaskName    string `json:"task_name"`
	TaskContent string `json:"task_content"`
	Cid         string `json:"cid"`
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
