package middleware

import (
	"bytes"
	"encoding/json"
	"github.com/dzjyyds666/opensource/logx"
	"github/dzjyyds666/online_study_server/user/api/internal/types"
	"net/http"
	"time"
)

// 格式化响应结果
// 自定义httprecoder，实现httpwriter接口，
// 在向httprespose写入数据的时候，记录要写入的数据，然后在中间件中拿到记录的数据，对数据进行处理，
// 最后重新写入到httprespose中

func FormatHttpResponse(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// 创建一个记录器响应捕获
		recorder := NewResponseRecorder(w)

		next(recorder, r)

		body := recorder.Body

		var httpRes types.HttpResponse
		err := json.Unmarshal(body.Bytes(), &httpRes)
		if err != nil {
			logx.GetLogger("OS_Server").Errorf("FormatHttpResponse|UnmarshalError|%v", err)
			httpRes.Code = 500
			httpRes.Msg = "服务器内部错误"
			httpRes.Time = time.Now().Unix()
			httpRes.RequestUrl = r.URL.String()
			httpRes.Data = nil
			b, _ := json.Marshal(httpRes)
			_, err := w.Write(b)
			if err != nil {
				logx.GetLogger("OS_Server").Errorf("FormatHttpResponse|ResponseRecorder Write err:%v", err)
			}
		} else {
			httpRes.Time = time.Now().Unix()
			httpRes.RequestUrl = r.URL.String()
			b, _ := json.Marshal(httpRes)
			_, err := w.Write(b)
			if err != nil {
				logx.GetLogger("OS_Server").Errorf("FormatHttpResponse|ResponseRecorder Write err:%v", err)
			}
		}
	}
}

// ResponseRecorder 用于捕获 HTTP 响应体
type ResponseRecorder struct {
	http.ResponseWriter
	Body *bytes.Buffer
}

// NewResponseRecorder 返回一个新的 ResponseRecorder
func NewResponseRecorder(w http.ResponseWriter) *ResponseRecorder {
	return &ResponseRecorder{
		ResponseWriter: w,
		Body:           new(bytes.Buffer),
	}
}

// Write 重写 Write 方法，以便捕获响应内容
func (rec *ResponseRecorder) Write(p []byte) (n int, err error) {
	// 将数据写入到 Body
	return rec.Body.Write(p)
}
