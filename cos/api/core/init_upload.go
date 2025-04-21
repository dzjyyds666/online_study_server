package core

import (
	"encoding/json"
)

var InitStatus = struct {
	Ing  string `json:"ing"`
	Done string `json:"done"`
	Fail string `json:"fail"`
}{
	Ing:  "ing",
	Done: "done",
	Fail: "fail",
}

type InitMultipartUpload struct {
	Fid        string `json:"fid,omitempty"`
	PartBytes  int64  `json:"part_bytes,omitempty"`
	TotalParts int64  `json:"total_parts,omitempty"`
	LastParts  int64  `json:"last_parts,omitempty"`
	UploadId   string `json:"upload_id,omitempty"`
	Status     string `json:"status,omitempty"` // 状态
}

func (imu *InitMultipartUpload) WithFid(fid string) *InitMultipartUpload {
	imu.Fid = fid
	return imu
}

func (imu *InitMultipartUpload) WithUploadId(uploadId string) *InitMultipartUpload {
	imu.UploadId = uploadId
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

func (imu *InitMultipartUpload) Marshal() string {
	data, _ := json.Marshal(imu)
	return string(data)
}

func (imu *InitMultipartUpload) WithStatus(status string) *InitMultipartUpload {
	imu.Status = status
	return imu
}
