package core

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/dzjyyds666/opensource/logx"
	"github.com/redis/go-redis/v9"
)

type TaskServer struct {
	classDB *redis.Client
	ctx     context.Context
}

func NewTaskServer(ctx context.Context, classDB *redis.Client) *TaskServer {
	return &TaskServer{
		ctx:     ctx,
		classDB: classDB,
	}
}

func (ts *TaskServer) CreateTask(ctx context.Context, task *Task) error {
	key := BuildTaskInfo(task.TaskId)
	err := ts.classDB.Set(ctx, key, task.Marshal(), 0).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("ClassServer|CreateTask|SetTaskInfoError|%v", err)
		return err
	}
	return err
}

func (ts *TaskServer) QueryTaskInfo(ctx context.Context, taskId string) (*Task, error) {
	key := BuildTaskInfo(taskId)
	taskInfo, err := ts.classDB.Get(ctx, key).Result()
	if err != nil {
		logx.GetLogger("study").Errorf("ClassServer|QueryTaskInfo|GetTaskInfoError|%v", err)
		return nil, err
	}
	var info *Task
	err = json.Unmarshal([]byte(taskInfo), &info)
	if err != nil {
		logx.GetLogger("study").Errorf("ClassServer|QueryTaskInfo|UnmarshalTaskInfoError|%v", err)
		return nil, err
	}
	return info, nil
}

func (ts *TaskServer) DeleteTask(ctx context.Context, taskId string) error {
	key := BuildTaskInfo(taskId)
	err := ts.classDB.Del(ctx, key).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("ClassServer|DeleteTask|DeleteTaskError|%v", err)
		return err
	}
	return nil
}

func (ts *TaskServer) UpdateTask(ctx context.Context, task *Task) error {
	// 先校验一下参数是否正常
	if len(task.TaskId) <= 0 || len(task.Cid) <= 0 || len(task.TaskName) <= 0 || len(task.TaskContent) <= 0 {
		return errors.New("params error")
	}
	key := BuildTaskInfo(task.TaskId)
	err := ts.classDB.Set(ctx, key, task.Marshal(), 0).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("ClassServer|CreateTask|SetTaskInfoError|%v", err)
		return err
	}
	return nil
}
