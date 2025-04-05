package core

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/dzjyyds666/opensource/common"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/dzjyyds666/opensource/sdk"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"user/api/config"
	mymiddleware "user/api/middleware"
)

type UserServer struct {
	ctx      context.Context
	dsClient *redis.Client
	mySql    *gorm.DB
}

func NewUserServer(ctx context.Context, dsClient *redis.Client, mySql *gorm.DB) *UserServer {
	return &UserServer{
		ctx:      ctx,
		dsClient: dsClient,
		mySql:    mySql,
	}
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
	var user *UserInfo
	err := us.mySql.WithContext(ctx).Where("uid = ?", uid).First(user).Error
	if err != nil {
		logx.GetLogger("study").Errorf("UserServer|QueryUserInfo|Query User Info Error|%v", err)
		return nil, err
	}

	return user, nil
}

func (us *UserServer) Login(ctx context.Context, uid, password string) error {
	var user UserInfo
	err := us.mySql.Where("uid = ?", uid).First(&user).Error
	if err != nil {
		logx.GetLogger("study").Errorf("UserServer|Login|Query User Info Error|%v", err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotExist
		} else {
			return err
		}
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		logx.GetLogger("study").Errorf("UserServer|Login|Password Error|%v", err)
		return ErrPasswordNotMatch
	}

	return nil
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
