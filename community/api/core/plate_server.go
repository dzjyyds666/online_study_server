package core

import (
	"context"
	"errors"
	"github.com/dzjyyds666/opensource/common"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

var lg = logx.GetLogger("study")

type PlateServer struct {
	ctx       context.Context
	rsDb      *redis.Client
	plateMgDb *mongo.Collection
}

func NewPlateServer(ctx context.Context, rsDb *redis.Client, mgDb *mongo.Client) *PlateServer {
	collection := mgDb.Database("learnX").Collection("plate")
	return &PlateServer{
		ctx:       ctx,
		rsDb:      rsDb,
		plateMgDb: collection,
	}
}

func (p *PlateServer) CreatePlate(ctx context.Context, plate *Plate) error {
	id := newPlateId(8)
	plate.WithId(id).
		WithCreateTime(time.Now().Unix()).
		WithArticleNumber(0).
		WithSubscribeNumber(0)

	key := buildPlateListKey()
	err := p.rsDb.ZAdd(ctx, key, redis.Z{
		Member: plate.Id,
		Score:  float64(time.Now().Unix()),
	}).Err()
	if err != nil {
		lg.Errorf("CreatePlate|ZAdd err:%v", err)
		return err
	}

	// 把信息存储到mongodb中
	_, err = p.plateMgDb.InsertOne(ctx, plate)
	if err != nil {
		lg.Errorf("CreatePlate|InsertOne err:%v", err)
		return err
	}
	lg.Infof("CreatePlate|CreatePlateSuccess|%v", common.ToStringWithoutError(plate))
	return nil
}

func (p *PlateServer) UpdatePlate(ctx context.Context, plate *Plate) error {
	// 更新mongodb的数据
	filter := bson.M{
		"_id": plate.Id,
	}
	update := bson.M{
		"$set": plate,
	}
	result, err := p.plateMgDb.UpdateOne(ctx, filter, update)
	if err != nil {
		lg.Errorf("UpdatePlate|UpdateOne err:%v", err)
		return err
	}

	if result.MatchedCount == 0 {
		lg.Errorf("UpdatePlate|No Data Match|%v", common.ToStringWithoutError(result))
		return ErrNoMatchData
	}

	lg.Errorf("UpdatePlate|UpdatePlateSuccess|%v", common.ToStringWithoutError(result))
	return nil
}

func (p *PlateServer) ListPlate(ctx context.Context) ([]*Plate, error) {
	key := buildPlateListKey()
	result, err := p.rsDb.ZRange(ctx, key, 0, -1).Result()
	if err != nil {
		lg.Errorf("ListPlate|ZRange err:%v", err)
		return nil, err
	}
	lg.Infof("ListPlate|ZRangeSuccess|%v", common.ToStringWithoutError(result))
	plates := make([]*Plate, 0)
	for _, id := range result {
		plate := &Plate{}
		err = p.plateMgDb.FindOne(ctx, bson.M{"_id": id}).Decode(plate)
		if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			lg.Errorf("ListPlate|FindOne err:%v", err)
			return nil, err
		} else if errors.Is(err, mongo.ErrNoDocuments) {
			lg.Errorf("ListPlate|No Data Match|%v", common.ToStringWithoutError(plate))
			continue
		}
		plates = append(plates, plate)
	}
	return plates, nil
}

func (p *PlateServer) QueryPlateInfo(ctx context.Context, pid string) (*Plate, error) {
	var plate Plate
	err := p.plateMgDb.FindOne(ctx, bson.M{"_id": pid}).Decode(&plate)
	if err != nil {
		lg.Errorf("QueryPlateInfo|FindOne err:%v", err)
		return nil, err
	}
	lg.Infof("QueryPlateInfo|QueryPlateInfoSuccess|%v", common.ToStringWithoutError(&plate))
	return &plate, nil
}
