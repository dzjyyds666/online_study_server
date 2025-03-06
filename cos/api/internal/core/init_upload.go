package core

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/labstack/echo"
)

type InitMultipartUpload struct {
	FileSize    int64  `json:"file_size,omitempty"`
	FileType    string `json:"file_type,omitempty"`
	DirectoryId string `json:"directory_id,omitempty"`
	Bucket      string `json:"bucket,omitempty"`
	DirId       string `json:"dir_id,omitempty"`
	Fid         string `json:"fid,omitempty"`
	PartBytes   int64  `json:"part_bytes,omitempty"`
	TotalParts  int64  `json:"total_parts,omitempty"`
	LastParts   int64  `json:"last_parts,omitempty"`
}

func (imu *InitMultipartUpload) WithFid(fid string) *InitMultipartUpload {
	imu.Fid = fid
	return imu
}

func (imu *InitMultipartUpload) WithFileSize(fileSize int64) *InitMultipartUpload {
	imu.FileSize = fileSize
	return imu
}

func (imu *InitMultipartUpload) WithFileType(fileType string) *InitMultipartUpload {
	imu.FileType = fileType
	return imu
}

func (imu *InitMultipartUpload) WithDirectoryId(directoryId string) *InitMultipartUpload {
	imu.DirectoryId = directoryId
	return imu
}

func (imu *InitMultipartUpload) WithBucket(bucket string) *InitMultipartUpload {
	imu.Bucket = bucket
	return imu
}

func (imu *InitMultipartUpload) WithDirId(dirId string) *InitMultipartUpload {
	imu.DirId = dirId
	return imu
}

func (imu *InitMultipartUpload) WithPartBytes(partBytes int64) *InitMultipartUpload {
	imu.PartBytes = partBytes
	return imu
}

func (imu *InitMultipartUpload) WithTotalParts(totalParts int64) *InitMultipartUpload {
	imu.TotalParts = totalParts
	return imu
}

func (imu *InitMultipartUpload) WithLastParts(lastParts int64) *InitMultipartUpload {
	imu.LastParts = lastParts
	return imu
}

func (imu *InitMultipartUpload) InitUpload(ctx echo.Context, client *s3.Client) error {
	return nil
}
