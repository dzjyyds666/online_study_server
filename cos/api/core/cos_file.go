package core

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"io"
	"path"
)

type CosFile struct {
	FileName    *string `json:"file_name,omitempty"`
	Fid         *string `json:"fid,omitempty"`
	FileMD5     *string `json:"file_md5,omitempty"`
	FileSize    *int64  `json:"file_size,omitempty"`
	FileType    *string `json:"file_type,omitempty"`
	DirectoryId *string `json:"directory_id,omitempty"`
}

func (cf *CosFile) MergeFilePath() string {
	return fmt.Sprintf("/%s/%s/%s%s", *cf.DirectoryId, *cf.Fid, *cf.Fid, path.Ext(*cf.FileName))
}

func (cf *CosFile) WithFileName(fileName string) *CosFile {
	cf.FileName = aws.String(fileName)
	return cf
}

func (cf *CosFile) WithFid(fid string) *CosFile {
	cf.Fid = aws.String(fid)
	return cf
}

func (cf *CosFile) WithFileMD5(fileMD5 string) *CosFile {
	cf.FileMD5 = aws.String(fileMD5)
	return cf
}

func (cf *CosFile) WithFileSize(fileSize int64) *CosFile {
	cf.FileSize = aws.Int64(fileSize)
	return cf
}

func (cf *CosFile) WithFileType(fileType string) *CosFile {
	cf.FileType = aws.String(fileType)
	return cf
}

func (cf *CosFile) WithDirectoryId(directoryId string) *CosFile {
	cf.DirectoryId = aws.String(directoryId)
	return cf
}

func (cf *CosFile) Marshal() string {
	marshal, _ := json.Marshal(cf)
	return string(marshal)
}

// todo 计算文件的md5
func CalculateMD5(reader io.Reader) (string, error) {
	buffer := make([]byte, 1024)
	md5Hash := md5.New()
	for {
		n, err := reader.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		md5Hash.Write(buffer[:n])
	}
	return hex.EncodeToString(md5Hash.Sum(nil)), nil
}

func GetFileType(reader io.Reader) (string, error) {
	buffer := make([]byte, 512)
	n, err := reader.Read(buffer)
	if err != nil {
		return "", err
	}

	detect := mimetype.Detect(buffer[:n])

	return detect.String(), nil
}

func GenerateFid() string {
	u := uuid.New()
	return "fi_" + u.String()
}
