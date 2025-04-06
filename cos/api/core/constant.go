package core

import "fmt"

const (
	RedisPrepareInfoKey  = "cos:%s:prepare:info"
	RedisInitInfoKey     = "cos:%s:init:info"
	RedisUploadPartIdKey = "cos:%s:upload:part:list"
	RedisInfoKey         = "cos:%s:info"
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

func buildUploadPartIdKey(fid string) string {
	return fmt.Sprintf(RedisUploadPartIdKey, fid)
}

var (
	ErrUploadVideoPart = fmt.Errorf("UploadVideoPart Error")
	ErrPartNotEnough   = fmt.Errorf("Part Not Enough")
)
