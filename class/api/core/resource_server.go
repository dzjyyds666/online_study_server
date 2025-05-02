package core

import (
	"common/proto"
	"common/rpc/client"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/dzjyyds666/opensource/logx"
	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
)

type ResourceServer struct {
	resourceDB *redis.Client
	ctx        context.Context
	mgCli      *mongo.Collection
}

func NewResourceServer(ctx context.Context, dsClient *redis.Client, mgCli *mongo.Client) *ResourceServer {
	return &ResourceServer{
		resourceDB: dsClient,
		ctx:        ctx,
		mgCli:      mgCli.Database("learnX").Collection("resource"),
	}
}

func (rs *ResourceServer) QueryResourceList(ctx context.Context, list *ResourceList) error {
	result, err := rs.resourceDB.ZRange(ctx, BuildChapterResourceList(list.SourceId), 0, -1).Result()
	if err != nil {
		logx.GetLogger("study").Errorf("ResourceServer|QueryResourceList|QueryResourceListError|%v", err)
		return err
	}
	for _, id := range result {
		resourceInfo, err := rs.QueryResourceInfo(ctx, id)
		if err != nil {
			logx.GetLogger("study").Errorf("ResourceServer|QueryResourceList|QueryResourceInfoError|%v", err)
			return err
		}
		list.ResourceList = append(list.ResourceList, *resourceInfo)
	}
	return nil
}

func (rs *ResourceServer) CreateResource(ctx context.Context, info *Resource) error {
	one, err := rs.mgCli.InsertOne(ctx, info)
	if err != nil {
		lg.Errorf("ResourceServer|CreateResource|CreateResource Error|%v", err)
		return err
	}
	if one.InsertedID == nil {
		lg.Errorf("ResourceServer|CreateResource|Insert Info Exist")
		return nil
	}
	return nil
}

func (rs *ResourceServer) DeleteResource(ctx context.Context, fid string) error {
	// 删除章节的info
	one, err := rs.mgCli.DeleteOne(ctx, bson.M{
		"_id": fid,
	})
	if err != nil {
		lg.Errorf("ResourceServer|DeleteResource|DeleteOne Error|%v", err)
		return err
	}

	if one.DeletedCount == 0 {
		lg.Errorf("ResourceServer|DeleteResource|No Such Resource")
		return nil
	}
	// todo 调用rpc删除cos的资源
	return nil
}

func (rs *ResourceServer) QueryResourceInfo(ctx context.Context, fid string) (*Resource, error) {

	cosRpcClient := client.GetCosRpcClient(ctx)
	info, err := cosRpcClient.GetFileInfo(ctx, &proto.ResourceInfo{
		Fid: fid,
	})
	if err != nil {
		logx.GetLogger("study").Errorf("ResourceServer|QueryResourceInfo|GetFileInfo Error|%v", err)
		return nil, err
	}

	var resource Resource
	err = rs.mgCli.FindOne(ctx, bson.M{
		"_id": fid,
	}).Decode(&resource)
	if err != nil {
		lg.Errorf("ResourceServer|QueryResourceInfo|FindOne Error|%v", err)
		return nil, err
	}
	resource.WithFileInfo(info)
	return &resource, nil
}

// 更新资源的状态
func (rs *ResourceServer) UpdateResource(ctx context.Context, info *Resource) error {
	id, err := rs.mgCli.UpdateByID(ctx, info.Fid, bson.M{
		"$set": info,
	})

	if err != nil {
		lg.Errorf("ResourceServer|UpdateResource|Update Resource Error|%v", err)
		return err
	}
	if id.ModifiedCount == 0 {
		lg.Errorf("ResourceServer|UpdateResource|Update resource fail")
	}
	return nil
}
