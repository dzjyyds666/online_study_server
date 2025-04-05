package core

import "errors"

const (
	RedisTokenKey      = "user.proto:login:token:%s"
	RedisVerifyCodeKey = "user.proto:verify_code:%s"
)

var (
	ErrPasswordNotMatch = errors.New("密码错误")
	ErrUserNotExist     = errors.New("用户不存在")
)
