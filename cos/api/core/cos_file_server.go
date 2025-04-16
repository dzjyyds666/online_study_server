package core

import (
	"context"
	"cos/api/config"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"

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

func (cfs *CosFileServer) PushVideoToLambdaQueue(ctx context.Context, fid string) error {
	err := cfs.cosDB.LPush(ctx, buildVideoLambdaQueueKey(), fid).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("PushVideoToLambdaQueue|LPush Error|%v", err)
		return err
	}
	return nil
}

func (cfs *CosFileServer) GetFile(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	object, err := cfs.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		logx.GetLogger("study").Errorf("GetFile|GetObject Error|%v", err)
		return nil, err
	}
	return object.Body, nil
}

func (cfs *CosFileServer) CheckFile(ctx context.Context, fid string) (io.ReadCloser, *string, error) {
	file, err := cfs.QueryCosFile(ctx, fid)
	if err != nil {
		logx.GetLogger("study").Errorf("CheckFile|QueryCosFile Error|%v", err)
		return nil, nil, err
	}
	r, err := cfs.GetFile(ctx, cfs.bucket, file.MergeFilePath())
	if err != nil {
		logx.GetLogger("study").Errorf("CheckFile|GetFile Error|%v", err)
		return nil, nil, err
	}
	return r, file.FileType, nil
}

func (cfs *CosFileServer) DeleteFile(ctx context.Context, fid string) error {
	info, err := cfs.QueryCosFile(ctx, fid)
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
