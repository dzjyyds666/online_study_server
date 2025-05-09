package core

import (
	"context"
	"cos/api/config"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/dzjyyds666/opensource/common"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/redis/go-redis/v9"
)

type CosFileServer struct {
	cosDB    *redis.Client
	s3Client *s3.Client
	ctx      context.Context
	bucket   string
}

func NewCosFileServer(ctx context.Context, cosDB *redis.Client, s3Client *s3.Client) *CosFileServer {
	return &CosFileServer{
		cosDB:    cosDB,
		ctx:      ctx,
		s3Client: s3Client,
		bucket:   *config.GloableConfig.S3.Bucket[0],
	}
}

func (cfs *CosFileServer) QueryCosFile(ctx context.Context, fid string) (*CosFile, error) {
	key := buildFileInfoKey(fid)
	result, err := cfs.cosDB.Get(ctx, key).Result()
	if err != nil {
		logx.GetLogger("study").Errorf("CosFileServer|QueryCosFile|GetCosFileInfoError|%v", err)
		return nil, err
	}

	var cosFile CosFile
	err = json.Unmarshal([]byte(result), &cosFile)
	if err != nil {
		logx.GetLogger("study").Errorf("CosFileServer|QueryCosFile|UnmarshalError|%v", err)
		return nil, err
	}
	return &cosFile, nil
}

func (cfs *CosFileServer) SaveFileInfo(ctx context.Context, cosFile *CosFile) error {
	key := buildFileInfoKey(*cosFile.Fid)
	err := cfs.cosDB.Set(ctx, key, cosFile.Marshal(), 0).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("CosFileServer|CopyNewFile|SetError|%v", err)
		return err
	}
	return nil
}

func (cfs *CosFileServer) ApplyUpload(ctx context.Context, cosFile *CosFile) error {
	key := buildPrepareFileInfoKey(*cosFile.Fid)
	err := cfs.cosDB.Set(ctx, key, cosFile.Marshal(), 0).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("CosFileServer|ApplyUpload|SetError|%v", err)
		return err
	}
	return nil
}

func (cfs *CosFileServer) CheckBucketExist() error {
	for _, bucket := range config.GloableConfig.S3.Bucket {
		_, err := cfs.s3Client.HeadBucket(context.TODO(), &s3.HeadBucketInput{
			Bucket: bucket,
		})
		if err != nil {
			logx.GetLogger("study").Errorf("CheckAndCreateBucket|HeadBucket err:%v", err)

			// 创建bucket
			_, err = cfs.s3Client.CreateBucket(context.TODO(), &s3.CreateBucketInput{
				Bucket: bucket,
			})
			if err != nil {
				logx.GetLogger("study").Errorf("CheckAndCreateBucket|CreateBucket err:%v", err)
				return err
			}

			// 设置桶的策略
			policy := `
				{
     "Version": "2012-10-17",
     "Statement": [
       {
         "Effect": "Allow",
         "Principal": "*",
         "Action": "s3:GetObject",
         "Resource": "arn:aws:s3:::%s/*"
       }
     ]
   }
`
			policy = fmt.Sprintf(policy, *bucket)
			_, err = cfs.s3Client.PutBucketPolicy(context.TODO(), &s3.PutBucketPolicyInput{
				Bucket: bucket,
				Policy: &policy,
			})
			if err != nil {
				logx.GetLogger("study").Errorf("CheckAndCreateBucket|PutBucketPolicy err:%v", err)
				return err
			}
		}
	}
	return nil
}

func (cfs *CosFileServer) InitUpload(ctx context.Context, bucket string, init *InitMultipartUpload) error {
	info, err := cfs.QueryFilePrepareInfo(ctx, init.Fid)
	if err != nil {
		logx.GetLogger("study").Errorf("InitUpload|QueryFilePrepareInfo err:%v", err)
		return err
	}
	logx.GetLogger("study").Infof("InitUpload|QueryFilePrepareInfo|%v", common.ToStringWithoutError(info))
	//if strings.Contains(*info.FileType, "video") {
	//	err := cfs.InitUploadVideo(ctx, init)
	//	if err != nil {
	//		return err
	//	}
	//	return nil
	//} else {
	objectKey := info.MergeFilePath()
	upload, err := cfs.s3Client.CreateMultipartUpload(ctx, &s3.CreateMultipartUploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		logx.GetLogger("study").Errorf("InitUpload|CreateMultipartUpload err:%v", err)
		return err
	}
	init.WithUploadId(*upload.UploadId).WithStatus(InitStatus.Ing)
	// 把uplaod保存到redis中
	err = cfs.cosDB.Set(ctx, buildInitFileInfoKey(init.Fid), init.Marshal(), 0).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("InitUpload|SetError|%v", err)
		return err
	}
	logx.GetLogger("study").Infof("InitUpload|CreateMultipartUpload|%v", common.ToStringWithoutError(upload))
	return err
}

func (cfs *CosFileServer) InitUploadVideo(ctx context.Context, init *InitMultipartUpload) error {
	rawData := init.WithStatus(InitStatus.Ing).Marshal()
	err := cfs.cosDB.Set(ctx, buildInitFileInfoKey(init.Fid), rawData, 0).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("InitUpload|SetError|%v", err)
		return err
	}
	return nil
}

func (cfs *CosFileServer) QueryFilePrepareInfo(ctx context.Context, fid string) (*CosFile, error) {
	key := buildPrepareFileInfoKey(fid)
	result, err := cfs.cosDB.Get(ctx, key).Result()
	if err != nil {
		logx.GetLogger("study").Errorf("QueryFilePrepareInfo|Get Error|%v", err)
		return nil, err
	}
	var cosFile CosFile
	err = json.Unmarshal([]byte(result), &cosFile)
	if err != nil {
		logx.GetLogger("study").Errorf("QueryFilePrepareInfo|Unmarshal Error|%v", err)
		return nil, err
	}

	return &cosFile, nil
}

//func (cfs *CosFileServer) UploadVideoPart(ctx context.Context, fid string, partId string, reader io.Reader) error {
//	// 先存入本地的临时目录中
//	info, err := cfs.QueryFilePrepareInfo(ctx, fid)
//	if err != nil {
//		logx.GetLogger("study").Errorf("UploadVideoPart|QueryFilePrepareInfo Error|%v", err)
//		cfs.UpdateInitStatus(ctx, fid, InitStatus.Fail)
//		return err
//	}
//	tmpPath := fmt.Sprintf("%s/%s", config.GloableConfig.TmpDir, *info.Fid+partId+path.Ext(*info.FileName))
//
//	// 写入文件中
//	file, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE, 0666)
//	if err != nil {
//		logx.GetLogger("study").Errorf("UploadVideoPart|OpenFile Error|%v", err)
//		cfs.UpdateInitStatus(ctx, fid, InitStatus.Fail)
//		return err
//	}
//
//	defer file.Close()
//	_, err = io.Copy(file, reader)
//	if err != nil {
//		logx.GetLogger("study").Errorf("UploadVideoPart|io.Copy Error|%v", err)
//		cfs.UpdateInitStatus(ctx, fid, InitStatus.Fail)
//		return err
//	}
//	// 把分片id存入redis中
//	err = cfs.cosDB.SAdd(ctx, buildUploadPartIdKey(fid), partId).Err()
//	if err != nil {
//		logx.GetLogger("study").Errorf("UploadVideoPart|SAdd Error|%v", err)
//		cfs.UpdateInitStatus(ctx, fid, InitStatus.Fail)
//		return err
//	}
//	return nil
//}

//func (cfs *CosFileServer) UpdateInitStatus(ctx context.Context, fid string, status string) {
//	key := buildInitFileInfoKey(fid)
//	result, err := cfs.cosDB.Get(ctx, key).Result()
//	if err != nil {
//		logx.GetLogger("study").Errorf("UpdateInitStatus|Get Error|%v", err)
//		return
//	}
//	var init InitMultipartUpload
//	err = json.Unmarshal([]byte(result), &init)
//	if err != nil {
//		logx.GetLogger("study").Errorf("UpdateInitStatus|Unmarshal Error|%v", err)
//		return
//	}
//	init.WithStatus(status)
//	err = cfs.cosDB.Set(ctx, key, init.Marshal(), 0).Err()
//	if err != nil {
//		logx.GetLogger("study").Errorf("UpdateInitStatus|Set Error|%v", err)
//		return
//	}
//	return
//}

func (cfs *CosFileServer) CompleteUploadVideo(ctx context.Context, fid string) error {
	initInfo, err := cfs.QueryInitInfo(ctx, fid)
	if err != nil && !errors.Is(err, redis.Nil) {
		logx.GetLogger("study").Errorf("CompleteUploadVideo|QueryInitInfo Error|%v", err)
		return err
	}

	if errors.Is(err, redis.Nil) {
		logx.GetLogger("study").Infof("CompleteUploadVideo|complete")
		return nil
	}

	if initInfo.Status == InitStatus.Fail {
		logx.GetLogger("study").Infof("CompleteUploadVideo|UploadVideoPart Fail")
		return ErrUploadVideoPart
	} else if initInfo.Status == InitStatus.Done {
		logx.GetLogger("study").Infof("CompleteUploadVideo|UploadVideoPart Done")
		return nil
	}

	info, err := cfs.QueryFilePrepareInfo(ctx, fid)
	if err != nil {
		logx.GetLogger("study").Errorf("CompleteUploadVideo|QueryFilePrepareInfo Error|%v", err)
		return err
	}

	// 查询分片是否全部上传完成
	partNumber, err := cfs.cosDB.SCard(ctx, buildUploadPartIdKey(fid)).Result()
	if err != nil {
		logx.GetLogger("study").Errorf("CompleteUploadVideo|SCard Error|%v", err)
		return err
	}

	if partNumber < initInfo.TotalParts {
		logx.GetLogger("study").Errorf("CompleteUploadVideo|SCard Error|%v", err)
		return ErrPartNotEnough
	}

	tmpPath := fmt.Sprintf("%s/%s", config.GloableConfig.TmpDir, *info.Fid+path.Ext(*info.FileName))

	file, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logx.GetLogger("study").Errorf("CompleteUploadVideo|OpenFile Error|%v", err)
		return err
	}

	defer file.Close()

	// 合并分片
	for i := 1; i <= int(partNumber); i++ {
		partId := strconv.Itoa(i)
		partPath := fmt.Sprintf("%s/%s", config.GloableConfig.TmpDir, *info.Fid+partId+path.Ext(*info.FileName))
		open, err := os.Open(partPath)
		if err != nil {
			logx.GetLogger("study").Errorf("CompleteUploadVideo|OpenFile Error|%v", err)
			return err
		}
		_, err = io.Copy(file, open)
		if err != nil {
			logx.GetLogger("study").Errorf("CompleteUploadVideo|io.Copy Error|%v", err)
			return err
		}
		open.Close()
		os.Remove(partPath)
	}

	err = cfs.CompleteFileIndex(ctx, fid)
	if err != nil {
		logx.GetLogger("study").Errorf("CompleteFileIndex|CompleteFileIndex Error|%v", err)
		return err
	}

	// 删除init信息和partlist
	err = cfs.cosDB.Del(ctx, buildInitFileInfoKey(fid), buildUploadPartIdKey(fid)).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("CompleteFileIndex|Del Error|%v", err)
	}
	// TODO 异步任务执行转码生成hls格式的视频，并上传到minio
	return nil
}

func (cfs *CosFileServer) CompleteFileIndex(ctx context.Context, fid string) error {
	prepareKey := buildPrepareFileInfoKey(fid)
	infoKey := buildFileInfoKey(fid)

	err := cfs.cosDB.Rename(ctx, prepareKey, infoKey).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("CompleteFileIndex|Rename Error|%v", err)
		return err
	}
	return nil
}

func (cfs *CosFileServer) QueryInitInfo(ctx context.Context, fid string) (*InitMultipartUpload, error) {
	key := buildInitFileInfoKey(fid)
	result, err := cfs.cosDB.Get(ctx, key).Result()
	if err != nil {
		logx.GetLogger("study").Errorf("QueryInitInfo|Get Error|%v", err)
		return nil, err
	}
	var init InitMultipartUpload
	err = json.Unmarshal([]byte(result), &init)
	if err != nil {
		logx.GetLogger("study").Errorf("QueryInitInfo|Unmarshal Error|%v", err)
		return nil, err
	}
	return &init, nil
}

func (cfs *CosFileServer) SingleUpload(ctx context.Context, bucket, fid string, reader io.Reader) (*CosFile, error) {
	info, err := cfs.QueryFilePrepareInfo(ctx, fid)
	if err != nil {
		logx.GetLogger("study").Errorf("SingleUpload|QueryFilePrepareInfo Error|%v", err)
		return nil, err
	}

	objectKey := info.MergeFilePath()

	_, err = cfs.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Body:          reader,
		Bucket:        aws.String(bucket),
		Key:           aws.String(objectKey),
		ContentLength: info.FileSize,
		ContentType:   info.FileType,
	})

	if err != nil {
		logx.GetLogger("study").Errorf("SingleUpload|PutObject Error|%v", err)
		return nil, err
	}
	err = cfs.CompleteFileIndex(ctx, fid)
	if err != nil {
		logx.GetLogger("study").Errorf("SingleUpload|CompleteFileIndex Error|%v", err)
		return nil, err
	}
	if strings.Contains(*info.FileType, "video") {
		go func() {
			ctx1 := context.Background()
			err = cfs.VideoProcessing(ctx1, fid)
			if err != nil {
				logx.GetLogger("study").Errorf("CompleteMultUpload|VideoProcessing Error|%v", err)
			}
		}()
	}
	return info, nil
}

func (cfs *CosFileServer) CompleteMultUpload(ctx context.Context, bucket, fid string, completeParts []types.CompletedPart) (*CosFile, error) {

	initInfo, err := cfs.QueryInitInfo(ctx, fid)
	if err != nil {
		logx.GetLogger("study").Errorf("CompleteMultUpload|QueryInitInfo Error|%v", err)
		return nil, err
	}

	info, err := cfs.QueryFilePrepareInfo(ctx, fid)
	if err != nil {
		logx.GetLogger("study").Errorf("CompleteMultUpload|QueryFilePrepareInfo Error|%v", err)
		return nil, err
	}
	_, err = cfs.s3Client.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(info.MergeFilePath()),
		UploadId: aws.String(initInfo.UploadId),
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: completeParts,
		},
	})
	if err != nil {
		logx.GetLogger("study").Errorf("CompleteMultUpload|CompleteMultipartUpload Error|%v", err)
		return nil, err
	}
	// 删除初始化上传的信息
	err = cfs.cosDB.Del(ctx, buildInitFileInfoKey(fid)).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("CompleteMultUpload|Del Error|%v", err)
		return nil, err
	}
	err = cfs.cosDB.Rename(ctx, buildPrepareFileInfoKey(fid), buildFileInfoKey(fid)).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("CompleteMultUpload|Rename Error|%v", err)
		return nil, err
	}
	logx.GetLogger("study").Infof("CompleteMultUpload|CompleteMultipartUpload Success")
	if strings.Contains(*info.FileType, "video") {
		go func() {
			ctx1 := context.Background()
			err = cfs.VideoProcessing(ctx1, fid)
			if err != nil {
				logx.GetLogger("study").Errorf("CompleteMultUpload|VideoProcessing Error|%v", err)
			}
		}()
	}
	return info, nil
}

func (cfs *CosFileServer) AbortUpload(ctx context.Context, bucket, fid string) error {
	initInfo, err := cfs.QueryInitInfo(ctx, fid)
	if err != nil {
		logx.GetLogger("study").Errorf("AbortUpload|QueryInitInfo Error|%v", err)
		return err
	}
	info, err := cfs.QueryFilePrepareInfo(ctx, fid)
	if err != nil {
		logx.GetLogger("study").Errorf("AbortUpload|QueryFilePrepareInfo Error|%v", err)
		return err
	}
	_, err = cfs.s3Client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(info.MergeFilePath()),
		UploadId: aws.String(initInfo.UploadId),
	})
	if err != nil {
		logx.GetLogger("study").Errorf("AbortUpload|AbortMultipartUpload Error|%v", err)
		return err
	}
	err = cfs.cosDB.Del(ctx, buildInitFileInfoKey(fid), buildPrepareFileInfoKey(fid)).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("AbortUpload|Del Error|%v", err)
		return err
	}
	logx.GetLogger("study").Errorf("AbortUpload|AbortMultipartUpload Success")
	return nil
}

func (cfs *CosFileServer) UploadPart(ctx context.Context, bucket, fid, partId string, reader io.Reader) (*string, error) {
	initInfo, err := cfs.QueryInitInfo(ctx, fid)
	if err != nil {
		logx.GetLogger("study").Errorf("UploadPart|QueryInitInfo Error|%v", err)
		return nil, err
	}
	info, err := cfs.QueryFilePrepareInfo(ctx, fid)
	if err != nil {
		logx.GetLogger("study").Errorf("UploadPart|QueryFilePrepareInfo Error|%v", err)
		return nil, err
	}
	partIdInt, _ := strconv.Atoi(partId)
	result, err := cfs.s3Client.UploadPart(ctx, &s3.UploadPartInput{
		Bucket:     aws.String(bucket),
		Key:        aws.String(info.MergeFilePath()),
		UploadId:   aws.String(initInfo.UploadId),
		PartNumber: aws.Int32(int32(partIdInt)),
		Body:       reader,
	})
	if err != nil {
		logx.GetLogger("study").Errorf("UploadPart|UploadPart Error|%v", err)
		return nil, err
	}
	logx.GetLogger("study").Infof("UploadPart|UploadPart Success")
	return result.ETag, nil
}

func (cfs *CosFileServer) UploadClassCover(ctx context.Context, filename, dirId, fileType, md5 string, size int64, reader io.Reader) (string, error) {
	fid := GenerateFid()
	var file CosFile
	file.WithFid(fid).
		WithFileSize(size).
		WithFileType(fileType).
		WithFileName(filename).
		WithFileMD5(md5).
		WithDirectoryId(dirId)
	// 先创建文件的prepare信息
	err := cfs.ApplyUpload(ctx, &file)
	if err != nil {
		logx.GetLogger("study").Errorf("UploadClassCover|ApplyUpload Error|%v", err)
		return "", err
	}
	// 直传
	_, err = cfs.SingleUpload(ctx, cfs.bucket, fid, reader)
	if err != nil {
		logx.GetLogger("study").Errorf("UploadClassCover|SingleUpload Error|%v", err)
		return "", err
	}
	return *file.Fid, nil
}

//func (cfs *CosFileServer) PushVideoToLambdaQueue(ctx context.Context, fid string) error {
//	return cfs.VideoProcessing(ctx, fid)
//}

func (cfs *CosFileServer) GetFile(ctx context.Context, bucket string, info *CosFile) (io.ReadCloser, string, error) {
	// 判断文件的类型
	key := info.MergeFilePath()
	fileType := *info.FileType
	if strings.Contains(*info.FileType, "video") {
		//  先判断视频的m3u8文件是否存在
		object, _ := cfs.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(info.MergeVideoPath()),
		})
		// 如果 object 为 nil，也可以认为文件不存在
		if object == nil {
			logx.GetLogger("study").Warnf("GetFile|HeadObject|File Not Found|Object is nil")
		} else {
			// 存在，使用m3u8文件
			key = info.MergeVideoPath()
			fileType = "application/x-mpegURL"
		}
	}

	logx.GetLogger("study").Errorf("GetFile|GetObject|%s", key)

	object, err := cfs.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		logx.GetLogger("study").Errorf("GetFile|GetObject Error|%v", err)
		return nil, "", err
	}
	return object.Body, fileType, nil
}

func (cfs *CosFileServer) CheckFile(ctx context.Context, fid string) (io.ReadCloser, *string, error) {
	file, err := cfs.QueryCosFile(ctx, fid)
	if err != nil {
		logx.GetLogger("study").Errorf("CheckFile|QueryCosFile Error|%v", err)
		return nil, nil, err
	}
	r, filetype, err := cfs.GetFile(ctx, cfs.bucket, file)
	if err != nil {
		logx.GetLogger("study").Errorf("CheckFile|GetFile Error|%v", err)
		return nil, nil, err
	}
	return r, &filetype, nil
}

func (cfs *CosFileServer) DeleteFile(ctx context.Context, fid string) error {
	info, err := cfs.QueryCosFile(ctx, fid)
	if err != nil {
		logx.GetLogger("study").Errorf("CheckFile|QueryCosFile Error|%v", err)
		return err
	}
	key := info.MergeFilePath()
	_, err = cfs.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(cfs.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		logx.GetLogger("study").Errorf("DeleteFile|DeleteObject Error|%v", err)
		return err
	}

	// 删除文件的index信息
	err = cfs.cosDB.Del(ctx, buildFileInfoKey(fid)).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("DeleteFile|Del Error|%v", err)
		return err
	}
	return nil
}

func (cfs *CosFileServer) CopyObject(ctx context.Context, fid string) (string, error) {
	info, err := cfs.QueryCosFile(ctx, fid)
	if err != nil {
		logx.GetLogger("study").Errorf("CopyObject|QueryCosFile Error|%v", err)
		return "", err
	}
	key := info.MergeFilePath()
	// 生成新的fid
	info.WithFid(GenerateFid())
	newKey := info.MergeFilePath()

	logx.GetLogger("study").Infof("CopyObject|CopyObject|%s|%s", key, newKey)
	_, err = cfs.s3Client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(cfs.bucket),
		CopySource: aws.String(cfs.bucket + key),
		Key:        aws.String(newKey),
	})
	if err != nil {
		logx.GetLogger("study").Errorf("CopyObject|CopyObject Error|%v", err)
		return "", err
	}

	// 创建新的文件信息
	err = cfs.SaveFileInfo(ctx, info)
	if err != nil {
		logx.GetLogger("study").Errorf("CopyObject|SaveFileInfo Error|%v", err)
		return "", err
	}
	return *info.Fid, nil
}

func (cfs *CosFileServer) VideoProcessing(ctx context.Context, fid string) error {
	// 查询 MinIO 文件元信息
	cosFile, err := cfs.QueryCosFile(ctx, fid)
	if err != nil {
		logx.GetLogger("study").Errorf("VideoProcessing|QueryCosFile Error|%v", err)
		return err
	}

	// 获取文件流
	file, _, err := cfs.GetFile(ctx, cfs.bucket, cosFile)
	if err != nil {
		logx.GetLogger("study").Errorf("VideoProcessing|GetFile Error|%v", err)
		return err
	}
	defer file.Close()

	// 创建临时保存路径
	filePath := "/tmp/ffmpeg/" + fid + path.Ext(*cosFile.FileName)
	outFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		logx.GetLogger("study").Errorf("VideoProcessing|OpenFile Error|%v", err)
		return err
	}
	defer outFile.Close() // 确保写入完自动关闭

	// 直接使用 io.Copy 是更高效和稳定的方式
	written, err := io.Copy(outFile, file)
	if err != nil {
		logx.GetLogger("study").Errorf("VideoProcessing|Copy Error|%v", err)
		return err
	}
	logx.GetLogger("study").Infof("VideoProcessing|File written|Size: %d bytes", written)

	resolutions := map[string]string{
		"360p":  "640x360",
		"480p":  "854x480",
		"720p":  "1280x720",
		"1080p": "1920x1080",
	}

	// 调用 FFmpeg 进行转码
	err = cfs.transcodeToHLS(filePath, "/tmp/ffmpeg/"+fid, resolutions, *cosFile.DirectoryId, *cosFile.Fid)
	if err != nil {
		logx.GetLogger("study").Errorf("VideoProcessing|TranscodeToHLS Error|%v", err)
		return err
	}
	err = cfs.uploadHlsVideo(ctx, "/tmp/ffmpeg/"+fid+"/", cosFile)
	if err != nil {
		logx.GetLogger("study").Errorf("VideoProcessing|uploadHlsVideo Error|%v", err)
		return err
	}
	// 处理完之后，把文件文件上传到minio中
	return nil
}

func (cfs *CosFileServer) transcodeToHLS(inputPath, outputDir string, resolutions map[string]string, dirid, fid string) error {
	for res, _ := range resolutions {
		err := createDirIfAbsent(outputDir + "/" + res)
		if err != nil {
			return err
		}
	}

	for key, res := range resolutions {
		outPath := filepath.Join(outputDir, key)
		os.MkdirAll(outPath, 0755)
		cmd := exec.Command("ffmpeg",
			"-i", inputPath,
			"-vf", "scale="+res,
			"-c:a", "aac", "-ar", "48000", "-b:a", "128k",
			"-c:v", "h264", "-profile:v", "main", "-crf", "20", "-g", "48",
			"-hls_time", "6", "-hls_playlist_type", "vod",
			"-f", "hls", filepath.Join(outPath, key+".m3u8"),
		)

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("转码 %s 失败: %w", key, err)
		}
	}

	//  生成 master.m3u8 播放列表
	masterPath := filepath.Join(outputDir, "master.m3u8")
	masterFile, err := os.Create(masterPath)
	if err != nil {
		return fmt.Errorf("无法创建 master.m3u8: %w", err)
	}
	defer masterFile.Close()

	masterFile.WriteString("#EXTM3U\n\n")
	for key, res := range resolutions {
		bandwidth := estimateBandwidth(res) // 自定义函数估算带宽
		masterFile.WriteString(fmt.Sprintf("#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=%s\n", bandwidth, res))
		// fixme 需要替换为服务器地址
		masterFile.WriteString(fmt.Sprintf("%s/%s/%s/%s/%s/%s.m3u8\n\n", "http://127.0.0.1:9000", cfs.bucket, dirid, fid, key, key))
	}
	return nil
}

func estimateBandwidth(res string) int {
	switch res {
	case "1920x1080":
		return 5000000
	case "1280x720":
		return 3000000
	case "854x480":
		return 1000000
	case "640x360":
		return 600000
	case "426x240":
		return 300000
	default:
		return 1500000 // 默认值
	}
}

func (cfs *CosFileServer) uploadHlsVideo(ctx context.Context, rootDir string, info *CosFile) error {
	// todo 上传文件的是否上传视频的master.m3u8文件
	//for key, _ := range resolutions {
	//	filePath := filepath.Join(rootDir, key)
	//	filepath.WalkDir(filePath, func(path string, d fs.DirEntry, err error) error {
	//		if err != nil {
	//			logx.GetLogger("study").Errorf("createDirIfAbsent|WalkDir|%v", err)
	//			return err
	//		}
	//		if !d.IsDir() {
	//			file, err := os.OpenFile(path, os.O_RDONLY, 0666)
	//			if err != nil {
	//				logx.GetLogger("study").Errorf("createDirIfAbsent|OpenFile|%v", err)
	//				return err
	//			}
	//			defer file.Close()
	//			// 构建key上传文件
	//			objectKey := filepath.Join(*info.DirectoryId, *info.Fid, key, filepath.Base(file.Name()))
	//			logx.GetLogger("study").Infof("createDirIfAbsent|createDirIfAbsent|%s", objectKey)
	//			// 调用s3上传
	//			_, err = cfs.s3Client.PutObject(ctx, &s3.PutObjectInput{
	//				Bucket: aws.String(cfs.bucket),
	//				Key:    aws.String(objectKey),
	//				Body:   file,
	//			})
	//			if err != nil {
	//				logx.GetLogger("study").Errorf("createDirIfAbsent|PutObject|%v", err)
	//				return err
	//			}
	//		}
	//		return nil
	//	})
	//}

	filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			logx.GetLogger("study").Errorf("createDirIfAbsent|WalkDir|%v", err)
			return err
		}
		if !d.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				logx.GetLogger("study").Errorf("createDirIfAbsent|OpenFile|%v", err)
				return err
			}
			defer file.Close()
			// 构建key上传文件
			key := strings.TrimPrefix(path, rootDir)
			objectKey := filepath.Join(*info.DirectoryId, *info.Fid, key)
			_, err = cfs.s3Client.PutObject(ctx, &s3.PutObjectInput{
				Bucket: aws.String(cfs.bucket),
				Key:    aws.String(objectKey),
				Body:   file,
			})
			if err != nil {
				logx.GetLogger("study").Errorf("createDirIfAbsent|PutObject|%v", err)
				return err
			}
		}
		return nil
	})
	return nil
}

func createDirIfAbsent(dir string) error {
	logx.GetLogger("study").Infof("createDirIfAbsent|createDirIfAbsent|%s", dir)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			logx.GetLogger("study").Errorf("createDirIfAbsent|createDirIfAbsent|%s", err)
			return err
		}
	}
	return nil
}
