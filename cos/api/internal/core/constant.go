package core

import "fmt"

const (
	RedisPrepareIndexKey = "cos:%s:prepare:index"
	RedisInitIndexKey    = "cos:%s:init:index"
	RedisIndexKey        = "cos:%s:index"
)

func buildFileIndexKey(fid string) string {
	return fmt.Sprintf(RedisIndexKey, fid)
}

func buildPrepareFileIndexKey(fid string) string {
	return fmt.Sprintf(RedisPrepareIndexKey, fid)
}

func buildInitFileIndexKey(fid string) string {
	return fmt.Sprintf(RedisInitIndexKey, fid)
}
