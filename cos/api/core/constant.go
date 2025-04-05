package core

import "fmt"

const (
	RedisPrepareInfoKey = "cos:%s:prepare:info"
	RedisInitInfoKey    = "cos:%s:init:info"
	RedisInfoKey        = "cos:%s:info"
)

func buildFileInfoKey(fid string) string {
	return fmt.Sprintf(RedisInfoKey, fid)
}

func buildPrepareFileInfoKey(fid string) string {
	return fmt.Sprintf(RedisPrepareInfoKey, fid)
}

func buildInitFileInfoKey(fid string) string {
	return fmt.Sprintf(RedisInitInfoKey, fid)
}
