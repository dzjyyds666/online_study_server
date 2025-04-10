package core

type Task struct {
	TaskId      string `json:"task_id"`
	TaskName    string `json:"task_name"`
	TaskContent string `json:"task_content"`
	Cid         string `json:"cid"`
}

type StudentTask struct {
}
