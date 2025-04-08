package service

import (
	"bytes"
	"common/proto"
	"context"
	"cos/api/core"
	"github.com/dzjyyds666/opensource/logx"
	"io"
)

type CosRpcServer struct {
	cosServer *core.CosFileServer
	proto.UnimplementedCosServer
}

func (cs *CosRpcServer) UploadClassFile(stream proto.Cos_UploadClassCoverServer) error {
	var filename string
	var dirId string
	var fileType string
	var md5 string
	var fileSize int64
	var fileData bytes.Buffer
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			logx.GetLogger("study").Errorf("UploadClassCover|Recv Error|%v", err)
			return err
		}

		if len(filename) <= 0 {
			filename = chunk.FileName
		}
		if len(dirId) <= 0 {
			dirId = chunk.DirectoryId
		}
		if len(fileType) <= 0 {
			fileType = chunk.FileType
		}
		if len(md5) <= 0 {
			md5 = chunk.Md5
		}
		if fileSize <= 0 {
			fileSize = chunk.FileSize
		}

		fileData.Write(chunk.Content)
	}
	fid, err := cs.cosServer.UploadClassCover(context.Background(), filename, dirId, fileType, md5, fileSize, &fileData)
	if err != nil {
		logx.GetLogger("study").Errorf("UploadClassCover|UploadClassCover Error|%v", err)
		return err
	}
	return stream.SendAndClose(&proto.UploadClassCoverResp{
		Fid: fid,
	})
}

func (cs *CosRpcServer) AddVideoToLambdaQueue(ctx context.Context, in *proto.VideoInfo) (*proto.CosCommonResponse, error) {
	// 把视频fid写入redis队列中
	err := cs.cosServer.PushVideoToLambdaQueue(ctx, in.Fid)
	if err != nil {
		logx.GetLogger("study").Errorf("AddVideoToLambdaQueue|PushVideoToLambdaQueue Error|%v", err)
		return &proto.CosCommonResponse{
			Success: false,
		}, err
	}
	return &proto.CosCommonResponse{
		Success: true,
	}, nil
}
