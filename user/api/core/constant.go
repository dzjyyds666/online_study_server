package core

import (
	"errors"
	"fmt"
)

const (
	RedisTokenKey         = "user:login:token:%s"
	RedisVerifyCodeKey    = "user:verify_code:%s"
	RedisStudentClassList = "user:%s:class:list"
)

var (
	ErrPasswordNotMatch = errors.New("密码错误")
	ErrUserNotExist     = errors.New("用户不存在")
)

func buildStudentClassListKey(uid string) string {
	return fmt.Sprintf(RedisStudentClassList, uid)
}
