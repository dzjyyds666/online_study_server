package core

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"time"
	"user/api/config"
	mymiddleware "user/api/middleware"

	"github.com/dzjyyds666/opensource/common"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/dzjyyds666/opensource/sdk"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserServer struct {
	ctx      context.Context
	dsClient *redis.Client
	mySql    *gorm.DB
}

func NewUserServer(ctx context.Context, dsClient *redis.Client, mySql *gorm.DB) (*UserServer, error) {
	err := mySql.AutoMigrate(&UserInfo{})
	if err != nil {
		logx.GetLogger("study").Errorf("UserServer|StartError|AutoMigrate|err:%v", err)
		return nil, err
	}
	return &UserServer{
		ctx:      ctx,
		dsClient: dsClient,
		mySql:    mySql,
	}, nil
}

func (us *UserServer) RegisterUser(ctx context.Context, user *UserInfo) error {
	err := us.mySql.WithContext(ctx).Create(user).Error
	if err != nil {
		logx.GetLogger("study").Errorf("UserServer|RegisterUser|Create User Error|%v", err)
		return err
	}
	return nil
}

func (us *UserServer) UpdateUserInfo(ctx context.Context, user *UserInfo) error {
	err := us.mySql.WithContext(ctx).Where("uid = ?", user.Uid).Updates(user).Error
	if err != nil {
		logx.GetLogger("study").Errorf("UserServer|UpdateUserInfo|Update User Info Error|%v", err)
		return err
	}
	return nil
}

func (us *UserServer) QueryUserInfo(ctx context.Context, uid string) (*UserInfo, error) {
	var user UserInfo
	err := us.mySql.WithContext(ctx).Where("uid = ?", uid).First(&user).Error
	if err != nil {
		logx.GetLogger("study").Errorf("UserServer|QueryUserInfo|Query User Info Error|%v", err)
		return nil, err
	}

	return &user, nil
}

func (us *UserServer) Login(ctx context.Context, uid, password string) (int, error) {
	var user UserInfo
	err := us.mySql.Where("uid = ?", uid).First(&user).Error
	if err != nil {
		logx.GetLogger("study").Errorf("UserServer|Login|Query User Info Error|%v", err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, ErrUserNotExist
		} else {
			return 0, err
		}
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		logx.GetLogger("study").Errorf("UserServer|Login|Password Error|%v", err)
		return 0, ErrPasswordNotMatch
	}

	return user.Role, nil
}

func (us *UserServer) CreateToken(ctx context.Context, uid string, role int) (string, error) {
	// 生成token
	tokenInfo := mymiddleware.Token{
		Uid:  uid,
		Role: role,
	}
	logx.GetLogger("study").Infof("HandlerLogin|Login Success|%s", common.ToStringWithoutError(tokenInfo))
	bytes, _ := json.Marshal(tokenInfo)
	token, err := sdk.CreateJwtToken(*config.GloableConfig.Jwt.Secretkey, bytes)
	if err != nil {
		logx.GetLogger("study").Errorf("HandlerLogin|CreateJwtToken Error|%v", err)
		return "", err
	}
	return token, nil
}

func (us *UserServer) BetchAddStudentToClass(ctx context.Context, cid string, r io.Reader) ([]string, error) {
	rows, err := parseExcel("学生名单", r)
	if err != nil {
		logx.GetLogger("study").Errorf("BetchAddStudentToClass|ParseExcel Error|%v", err)
		return nil, err
	}
	logx.GetLogger("study").Infof("BetchAddStudentToClass|ParseExcel Success|%v", common.ToStringWithoutError(rows))
	defaultPassword, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		logx.GetLogger("study").Errorf("BetchAddStudentToClass|GenerateFromPassword Error|%v", err)
		return nil, err
	}
	ids := make([]string, 0, len(rows))
	users := make([]UserInfo, 0, len(rows))
	for index, row := range rows {
		if index == 0 {
			continue
		}
		var user UserInfo
		user.WithUid(row[0])
		err := us.mySql.Where("uid = ?", user.Uid).First(&user).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			logx.GetLogger("study").Errorf("BetchAddStudentToClass|Query User Info Error|%v", err)
			continue
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			user.WithName(row[1]).
				WithCollage(row[2]).
				WithMajor(row[3]).
				WithRole(UserRole.Student).
				WithStatus(UserStatus.Active).
				WithCreateTs(time.Now().Unix()).
				WithUpdateTs(time.Now().Unix()).
				WithPassword(string(defaultPassword))
			err = us.mySql.WithContext(ctx).Create(&user).Error
			if err != nil {
				logx.GetLogger("study").Errorf("BetchAddStudentToClass|Create User Info Error|%v", err)
				continue
			}
		}
		// 插入到学生的班级列表中
		err = us.dsClient.ZAdd(ctx, buildStudentClassListKey(user.Uid), redis.Z{
			Member: cid,
			Score:  float64(time.Now().Unix()),
		}).Err()
		if err != nil {
			logx.GetLogger("study").Errorf("BetchAddStudentToClass|Add Student To Class Error|%v", err)
			break
		}
		users = append(users, user)
		ids = append(ids, user.Uid)
	}
	return ids, nil
}

func (us *UserServer) AddStudentToClass(ctx context.Context, cid, uid, name string) error {
	var user UserInfo
	logx.GetLogger("study").Infof("AddStudentToClass|succes|%s", uid)
	user.WithUid(uid)
	defaultPassword, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		logx.GetLogger("study").Errorf("BetchAddStudentToClass|GenerateFromPassword Error|%v", err)
		return err
	}
	err = us.mySql.Where("uid = ?", user.Uid).First(&user).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logx.GetLogger("study").Errorf("AddStudentToClass|Query User Info Error|%v", err)
		return err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		user.WithName(name).
			WithCollage("未知").
			WithMajor("未知").
			WithRole(UserRole.Student).
			WithStatus(UserStatus.Active).
			WithCreateTs(time.Now().Unix()).
			WithUpdateTs(time.Now().Unix()).
			WithPassword(string(defaultPassword))
		err = us.mySql.WithContext(ctx).Create(&user).Error
		if err != nil {
			logx.GetLogger("study").Errorf("AddStudentToClass|Create User Info Error|%v", err)
			return err
		}
	}
	err = us.dsClient.ZAdd(ctx, buildStudentClassListKey(user.Uid), redis.Z{
		Member: cid,
		Score:  float64(time.Now().Unix()),
	}).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("AddStudentToClass|Add Student To Class Error|%v", err)
		return err
	}
	return nil
}

func (us *UserServer) QueryStudentClassList(ctx context.Context, uid string) ([]string, error) {
	// 查询学生的班级列表
	classList, err := us.dsClient.ZRange(ctx, buildStudentClassListKey(uid), 0, -1).Result()
	if err != nil {
		logx.GetLogger("study").Errorf("QueryStudentClassList|Query Student Class List Error|%v", err)
		return nil, err
	}
	return classList, nil
}
