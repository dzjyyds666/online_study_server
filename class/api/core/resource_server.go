package core

import (
	"common/proto"
	"common/rpc/client"
	"encoding/json"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
)

type ResourceServer struct {
	resourceDB *redis.Client
	ctx        context.Context
}

func NewResourceServer(ctx context.Context, dsClient *redis.Client) *ResourceServer {
	return &ResourceServer{
		resourceDB: dsClient,
		ctx:        ctx,
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
	// 先存储资源的信息
	data := info.Marshal()
	err := rs.resourceDB.Set(ctx, BuildResourceInfo(*info.Fid), data, 0).Err()
	if err != nil {
		logx.GetLogger("study").Infof("ResourceServer|CreateResource|Set Resource Error|%v", err)
		return err
	}
	return nil
}

func (rs *ResourceServer) DeleteResource(ctx context.Context, fid string) error {
	// 删除章节的info
	return rs.resourceDB.Del(ctx, BuildResourceInfo(fid)).Err()
}

func (rs *ResourceServer) QueryResourceInfo(ctx context.Context, fid string) (*Resource, error) {
	key := BuildResourceInfo(fid)
	result, err := rs.resourceDB.Get(ctx, key).Result()
	if err != nil {
		logx.GetLogger("study").Errorf("ResourceServer|QueryResourceInfo|Get Resource Info Error|%v", err)
		return nil, err
	}

	cosRpcClient := client.GetCosRpcClient(ctx)
	info, err := cosRpcClient.GetFileInfo(ctx, &proto.ResourceInfo{
		Fid: fid,
	})
	if err != nil {
		logx.GetLogger("study").Errorf("ResourceServer|QueryResourceInfo|GetFileInfo Error|%v", err)
		return nil, err
	}

	var resource Resource
	err = json.Unmarshal([]byte(result), &resource)
	if err != nil {
		logx.GetLogger("study").Errorf("ResourceServer|QueryResourceInfo|Unmarshal Resource Info Error|%v", err)
		return nil, err
	}
	resource.WithFileInfo(info)
	return &resource, nil
}

// 更新资源的状态
func (rs *ResourceServer) UpdateResource(ctx context.Context, info *Resource) (*Resource, error) {
	// 更新resource的状态
	// 获取资源的信息
	resourceInfo, err := rs.QueryResourceInfo(ctx, *info.Fid)
	if err != nil {
		logx.GetLogger("study").Errorf("ResourceServer|UpdateResource|Query Resource Info Error|%v", err)
		return nil, err
	}
	if info.Downloadable != nil {
		resourceInfo.Downloadable = info.Downloadable
	}

	if info.Published != nil {
		resourceInfo.Published = info.Published
	}

	data := resourceInfo.Marshal()
	err = rs.resourceDB.Set(ctx, BuildResourceInfo(*info.Fid), data, 0).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("ResourceServer|UpdateResource|Set Resource Info Error|%v", err)
		return nil, err
	}
	return resourceInfo, nil
}
