package server

import (
	"bytes"
	"cos/api/internal/core"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"testing"
	"time"
)

func TestUploadFile(t *testing.T) {
	open, _ := os.Open("test.png")
	stat, _ := open.Stat()
	filename := stat.Name()
	fileSize := stat.Size()
	md5, _ := core.CalculateMD5(open)
	filetype := "image/png"
	dirid := "test"

	open.Seek(0, 0)

	var cosFile core.CosFile
	cosFile.WithFileName(filename).
		WithFileSize(fileSize).
		WithFileMD5(md5).
		WithFileType(filetype).
		WithDirectoryId(dirid)

	marshal, _ := json.Marshal(cosFile)

	t.Logf("UploadFile|cosFile:%v", string(marshal))

	buffer := bytes.NewBuffer(marshal)
	// 申请文件上传
	request, _ := http.NewRequest("POST", "http://localhost:19002/v1/api/cos/upload/apply", buffer)

	hcli := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, _ := hcli.Do(request)
	if resp.StatusCode != 200 {
		t.Errorf("UploadFile|err:%v", resp.StatusCode)
		return
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body) // 使用 io.ReadAll 读取响应体
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	t.Logf("UploadFile|body:%v", string(body))
	var infomap map[string]interface{}
	err = json.Unmarshal(body, &infomap)
	i, err := json.Marshal(infomap["data"])
	err = json.Unmarshal(i, &cosFile)
	t.Logf("UploadFile|cosFile:%v", aws.ToString(cosFile.Fid))

	// 上传文件
	// 定义form表单
	body1 := &bytes.Buffer{}
	writer := multipart.NewWriter(body1)
	h := textproto.MIMEHeader{}
	h.Set("Content-Disposition", "form-data; name=\"file\"; filename=\"test.png\"")
	part, _ := writer.CreatePart(h)
	io.Copy(part, open)

	writer.Close()

	newrequest, _ := http.NewRequest("POST", fmt.Sprintf("http://localhost:19002/v1/api/cos/upload/%s?dirid=%s",
		aws.ToString(cosFile.Fid), aws.ToString(cosFile.DirectoryId)), body1)
	newrequest.Header.Set("Content-Type", writer.FormDataContentType())
	do, err := hcli.Do(newrequest)
	if do.StatusCode != 200 {
		t.Errorf("UploadFile|err:%v", do.StatusCode)
		return
	}
	t.Logf("UploadFile|do:%v", do.StatusCode)
}

func TestGetFile(t *testing.T) {
	
}
