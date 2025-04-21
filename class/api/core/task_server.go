package core

import (
	"common/proto"
	"common/rpc/client"
	"context"
	"encoding/json"
	"errors"
	"github.com/dzjyyds666/opensource/common"
	"math"
	"strconv"
	"time"

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
	lg.Infof("ClassServer|QueryTaskInfo|QueryTaskInfo|%v", key)
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

func (ts *TaskServer) HasSubmit(ctx context.Context, taskId string, uid string) (bool, error) {
	key := BuildTaskStudentListKey(taskId)
	_, err := ts.classDB.ZScore(ctx, key, uid).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (ts *TaskServer) TaskSubmit(ctx context.Context, task *SubmitTask) error {
	// 生成taskid

	submit, err2 := ts.HasSubmit(ctx, task.TaskId, task.Owner)
	if err2 != nil {
		lg.Errorf("ClassServer|TaskSubmit|HasSubmitError|%v", err2)
		return err2
	}

	if submit {
		return ErrTaskHasSubmit
	}

	task.WithId(NewStudentTaskId(8)).WithViewing(false)
	raw, err := task.Marshal()
	if err != nil {
		lg.Errorf("ClassServer|TaskSubmit|MarshalError|%v", err)
		return err
	}
	// 存入task的信息
	key := BuildStudentTaskInfoKey(task.Id)
	err = ts.classDB.Set(ctx, key, raw, 0).Err()
	if err != nil {
		lg.Errorf("ClassServer|TaskSubmit|SetTaskInfoError|%v", err)
		return err
	}
	// 把作业信息存储到作业提交的学生列表下面
	key = BuildTaskStudentTaskListKey(task.TaskId)
	err = ts.classDB.ZAdd(ctx, key, redis.Z{
		Member: task.Id,
		Score:  float64(time.Now().Unix()),
	}).Err()
	if err != nil {
		lg.Errorf("ClassServer|TaskSubmit|AddTaskToStudentError|%v", err)
		return err
	}

	err = ts.classDB.ZAdd(ctx, BuildTaskStudentListKey(task.TaskId), redis.Z{
		Member: task.Owner,
		Score:  float64(time.Now().Unix()),
	}).Err()
	if err != nil {
		lg.Errorf("ClassServer|TaskSubmit|AddTaskToStudentError|%v", err)
		return err
	}

	// 存储到学生的作业列表下面
	return ts.classDB.ZAdd(ctx, BuildStudentTaskListKey(task.Owner), redis.Z{
		Member: task.Id,
		Score:  float64(time.Now().Unix()),
	}).Err()
}

func (ts *TaskServer) GetTaskListNumber(ctx context.Context, taskId string) (int64, error) {
	// 查询zset移动多少条数据
	result, err2 := ts.classDB.ZCard(ctx, BuildTaskStudentTaskListKey(taskId)).Result()
	if err2 != nil {
		lg.Errorf("ClassServer|ListStudentTask|GetTaskIdListError|%v", err2)
		return 0, err2
	}

	return result, err2
}

func (ts *TaskServer) ListStudentTask(ctx context.Context, list *ListStudentList) error {
	zrangeBy := &redis.ZRangeBy{
		Min:    "(0",
		Max:    strconv.FormatInt(math.MaxInt64, 10),
		Count:  list.Limit,
		Offset: (list.Page - 1) * list.Limit,
	}
	taskIds, err := ts.classDB.ZRangeByScore(ctx, BuildTaskStudentTaskListKey(list.TaskId), zrangeBy).Result()
	if err != nil {
		lg.Errorf("ClassServer|ListStudentTask|GetTaskIdListError|%v", err)
		return err
	}

	list.Tasks = make([]*SubmitTask, 0, len(taskIds))
	lg.Infof("ClassServer|ListStudentTask|GetTaskIdListSuccess|%s", common.ToStringWithoutError(taskIds))

	for _, taskId := range taskIds {
		task, err := ts.QueryStudentTask(ctx, taskId)
		if err != nil {
			lg.Errorf("ClassServer|ListStudentTask|GetTaskInfoError|%v", err)
			continue
		}
		list.Tasks = append(list.Tasks, task)
	}
	lg.Infof("ClassServer|ListStudentTask|ListStudentTaskSuccess|%s", common.ToStringWithoutError(list))
	return nil
}

func (ts *TaskServer) QueryStudentTask(ctx context.Context, taskId string) (*SubmitTask, error) {
	key := BuildStudentTaskInfoKey(taskId)
	result, err := ts.classDB.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	var st SubmitTask
	err = json.Unmarshal([]byte(result), &st)
	if err != nil {
		return nil, err
	}

	// 学生的信息
	rpcClient := client.GetUserRpcClient(ctx)
	info, err := rpcClient.GetStudentsInfo(ctx, &proto.StudentIds{Uids: []string{st.Owner}})
	if err != nil {
		return nil, err
	}

	st.OwnerName = info.Infos[0].Name
	return &st, nil
}

func (ts *TaskServer) UpdateStudentTask(ctx context.Context, task *SubmitTask) error {
	key := BuildStudentTaskInfoKey(task.Id)
	result, err := ts.classDB.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	var st SubmitTask
	err = json.Unmarshal([]byte(result), &st)
	if err != nil {
		return err
	}
	st.WithAnnotate(task.Annotate)
	st.WithLevel(task.Level)
	st.WithViewing(true)
	raw, err := st.Marshal()
	if err != nil {
		return err
	}
	return ts.classDB.Set(ctx, key, raw, 0).Err()
}

func (ts *TaskServer) ListOwnerTask(ctx context.Context, uid string) ([]*ListOwnerTask, error) {
	result, err := ts.classDB.ZRange(ctx, BuildStudentTaskListKey(uid), 0, -1).Result()
	if err != nil {
		return nil, err
	}

	lg.Infof("ClassServer|ListOwnerTask|GetTaskIdListSuccess|%s", common.ToStringWithoutError(result))

	list := make([]*ListOwnerTask, 0, len(result))

	for _, id := range result {
		task, err := ts.QueryStudentTask(ctx, id)
		if err != nil {
			return nil, err
		}
		var item ListOwnerTask
		item.Submit = task

		// 查对应的任务信息
		taskInfo, err := ts.QueryTaskInfo(ctx, task.TaskId)
		if err != nil {
			return nil, err
		}
		item.Task = taskInfo
		//查询课程信息
		result, err := ts.classDB.Get(ctx, BuildClassInfo(taskInfo.Cid)).Result()
		if err != nil {
			lg.Errorf("ClassServer|QueryClassInfo|GetClassInfoError|%v", err)
			return nil, err
		}

		var class Class
		err = json.Unmarshal([]byte(result), &class)
		if err != nil {
			lg.Errorf("ClassServer|QueryClassInfo|UnmarshalClassInfoError|%v", err)
			return nil, err
		}
		item.ClassInfo = &class
		list = append(list, &item)
	}
	return list, nil
}
