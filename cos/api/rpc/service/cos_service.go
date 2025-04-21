package service

import (
	"bytes"
	"common/proto"
	"context"
	"cos/api/core"
	"io"

	"github.com/dzjyyds666/opensource/logx"
)

type CosRpcServer struct {
	CosServer *core.CosFileServer
	proto.UnimplementedCosServer
}

func (cs *CosRpcServer) CopyObject(ctx context.Context, in *proto.CopyObjectRequest) (*proto.CopyObjectResponse, error) {
	fid, err := cs.CosServer.CopyObject(ctx, in.Fid)
	if err != nil {
		logx.GetLogger("study").Errorf("CopyObject|CopyObject Error|%v", err)
		return nil, err
	}
	return &proto.CopyObjectResponse{
		NewFid: fid,
	}, nil
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
	fid, err := cs.CosServer.UploadClassCover(context.Background(), filename, dirId, fileType, md5, fileSize, &fileData)
	if err != nil {
		logx.GetLogger("study").Errorf("UploadClassCover|UploadClassCover Error|%v", err)
		return err
	}
	return stream.SendAndClose(&proto.UploadClassCoverResp{
		Fid: fid,
	})
}

//func (cs *CosRpcServer) AddVideoToLambdaQueue(ctx context.Context, in *proto.VideoInfo) (*proto.CosCommonResponse, error) {
//	// 把视频fid写入redis队列中
//	err := cs.CosServer.PushVideoToLambdaQueue(ctx, in.Fid)
//	if err != nil {
//		logx.GetLogger("study").Errorf("AddVideoToLambdaQueue|PushVideoToLambdaQueue Error|%v", err)
//		return &proto.CosCommonResponse{
//			Success: false,
//		}, err
//	}
//	return &proto.CosCommonResponse{
//		Success: true,
//	}, nil
//}

func (cs *CosRpcServer) GetFileInfo(ctx context.Context, in *proto.ResourceInfo) (*proto.ResourceInfo, error) {
	file, err := cs.CosServer.QueryCosFile(ctx, in.GetFid())
	if err != nil {
		logx.GetLogger("study").Errorf("GetFileInfo|QueryCosFile Error|%v", err)
		return nil, err
	}
	info := &proto.ResourceInfo{
		Fid:      in.GetFid(),
		FileType: *file.FileType,
		FileSize: *file.FileSize,
		FileName: *file.FileName,
	}
	return info, nil
}

func (cs *CosRpcServer) DeleteTaskImage(ctx context.Context, in *proto.ImageIds) (*proto.CosCommonResponse, error) {
	fids := in.GetFids()
	logx.GetLogger("study").Errorf("DeleteTaskImage|fids|%v", fids)
	for _, fid := range fids {
		err := cs.CosServer.DeleteFile(ctx, fid)
		if err != nil {
			logx.GetLogger("study").Errorf("DeleteTaskImage|DeleteFile Error|%v", err)
		}
	}
	return &proto.CosCommonResponse{
		Success: true,
	}, nil
}
