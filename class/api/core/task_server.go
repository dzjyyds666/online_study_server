package core

import (
	"common/proto"
	"common/rpc/client"
	"context"
	"errors"
	"github.com/dzjyyds666/opensource/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"math"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type TaskServer struct {
	classDB    *redis.Client
	ctx        context.Context
	taskMgCli  *mongo.Collection
	subMgCli   *mongo.Collection
	classMgCli *mongo.Collection
}

func NewTaskServer(ctx context.Context, classDB *redis.Client, mongoCLi *mongo.Client) *TaskServer {
	return &TaskServer{
		ctx:        ctx,
		classDB:    classDB,
		taskMgCli:  mongoCLi.Database("learnX").Collection("task"),
		subMgCli:   mongoCLi.Database("learnX").Collection("task_submit"),
		classMgCli: mongoCLi.Database("learnX").Collection("class"),
	}
}

func (ts *TaskServer) CreateTask(ctx context.Context, task *Task) error {
	one, err := ts.taskMgCli.InsertOne(ctx, task)
	if err != nil {
		lg.Errorf("ClassServer|CreateTask|CreateTaskError|%v", err)
		return err
	}
	if one.InsertedID == nil {
		lg.Errorf("ClassServer|CreateTask|Insert Info Exist")
		return nil
	}
	return nil
}

func (ts *TaskServer) QueryTaskInfo(ctx context.Context, taskId string) (*Task, error) {
	var task Task
	err := ts.taskMgCli.FindOne(ctx, bson.M{"_id": taskId}).Decode(&task)
	if err != nil {
		lg.Errorf("ClassServer|QueryTaskInfo|QueryTaskInfoError|%v", err)
		return nil, err
	}
	return &task, nil
}

func (ts *TaskServer) DeleteTask(ctx context.Context, taskId string) error {
	one, err := ts.taskMgCli.DeleteOne(ctx, bson.M{"_id": taskId})
	if err != nil {
		lg.Errorf("ClassServer|DeleteTask|DeleteTaskError|%v", err)
		return err
	}
	if one.DeletedCount == 0 {
		lg.Errorf("ClassServer|DeleteTask|No Such Task")
		return nil
	}
	return nil
}

func (ts *TaskServer) UpdateTask(ctx context.Context, task *Task) error {
	// 先校验一下参数是否正常
	if len(task.TaskId) <= 0 || len(task.Cid) <= 0 || len(task.TaskName) <= 0 || len(task.TaskContent) <= 0 {
		return errors.New("params error")
	}
	id, err := ts.taskMgCli.UpdateByID(ctx, task.TaskId, bson.M{
		"$set": task,
	})
	if err != nil {
		lg.Errorf("ClassServer|UpdateTask|UpdateTaskError|%v", err)
		return err
	}
	if id.ModifiedCount == 0 {
		lg.Errorf("ClassServer|UpdateTask|Update Task Error|%v", err)
		return errors.New("updateError")
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
	submit, err := ts.HasSubmit(ctx, task.TaskId, task.Owner)
	if err != nil {
		lg.Errorf("ClassServer|TaskSubmit|HasSubmitError|%v", err)
		return err
	}

	if submit {
		return ErrTaskHasSubmit
	}

	lg.Infof("ClassServer|TaskSubmit|TaskSubmit|%s", common.ToStringWithoutError(task))

	task.WithId(NewStudentTaskId(8)).WithViewing(false)
	one, err := ts.subMgCli.InsertOne(ctx, task)
	if err != nil {
		lg.Errorf("ClassServer|TaskSubmit|CreateTaskError|%v", err)
		return err
	}

	if one.InsertedID == nil {
		lg.Errorf("ClassServer|TaskSubmit|Insert Info Exist")
		return nil
	}

	// 把作业信息存储到作业提交的学生列表下面
	key := BuildTaskStudentTaskListKey(task.TaskId)
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

func (ts *TaskServer) ListStudentTask(ctx context.Context, list *ListStudentTask) error {
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

	var st SubmitTask
	err := ts.subMgCli.FindOne(ctx, bson.M{"_id": taskId}).Decode(&st)
	if err != nil {
		return nil, err
	}
	// 学生的信息
	lg.Infof("ClassServer|QueryStudentTask|QueryStudentTaskSuccess|%s", common.ToStringWithoutError(st))
	rpcClient := client.GetUserRpcClient(ctx)
	info, err := rpcClient.GetStudentsInfo(ctx, &proto.StudentIds{Uids: []string{st.Owner}})
	if err != nil {
		return nil, err
	}
	st.OwnerName = info.Infos[0].Name
	return &st, nil
}

func (ts *TaskServer) UpdateStudentTask(ctx context.Context, task *SubmitTask) error {
	task.WithViewing(true)
	id, err := ts.subMgCli.UpdateByID(ctx, task.Id, bson.M{
		"$set": task,
	})
	if err != nil {
		lg.Errorf("ClassServer|UpdateStudentTask|UpdateStudentTaskError|%v", err)
		return err
	}
	if id.ModifiedCount == 0 {
		lg.Errorf("ClassServer|UpdateStudentTask|UpdateStudentTaskError|%v", err)
		return errors.New("updateError")
	}
	return nil
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
			lg.Errorf("ClassServer|ListOwnerTask|QueryStudentTaskError|%v", err)
			continue
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
		var class Class
		err = ts.classMgCli.FindOne(ctx, bson.M{
			"_id": taskInfo.Cid,
		}).Decode(&class)
		if err != nil {
			lg.Errorf("ClassServer|ListOwnerTask|GetClassInfoError|%v|%s", err, taskInfo.Cid)
			continue
		}
		item.ClassInfo = &class
		list = append(list, &item)
	}
	return list, nil
}
