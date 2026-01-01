package service

import (
	"context"

	v1 "smart-collab-gallery-server/api/file/v1"
	"smart-collab-gallery-server/internal/pkg"

	"github.com/go-kratos/kratos/v2/log"
)

type FileService struct {
	v1.UnimplementedFileServer

	cosManager *pkg.COSManager
	log        *log.Helper
}

func NewFileService(cosManager *pkg.COSManager, logger log.Logger) *FileService {
	return &FileService{
		cosManager: cosManager,
		log:        log.NewHelper(logger),
	}
}

// GetUploadPresignedUrl 获取上传预签名 URL（支持多存储桶）
func (s *FileService) GetUploadPresignedUrl(ctx context.Context, req *v1.GetUploadPresignedUrlRequest) (*v1.GetUploadPresignedUrlReply, error) {
	s.log.WithContext(ctx).Infof("获取上传预签名 URL: fileName=%s, contentType=%s, bucketKey=%s, fileSize=%d, autoDetect=%v",
		req.FileName, req.ContentType, req.BucketKey, req.FileSize, req.AutoDetect)

	// 检查 COS Manager 是否可用
	if s.cosManager == nil {
		s.log.WithContext(ctx).Error("COS Manager 未初始化，请检查配置")
		return nil, v1.ErrorSystemError("文件上传服务暂不可用，请联系管理员配置 COS")
	}

	// 参数校验
	if req.FileName == "" {
		return nil, v1.ErrorParamsError("文件名不能为空")
	}

	// 确定使用的存储桶 key
	bucketKey := req.BucketKey
	if bucketKey == "" {
		// 如果未指定且开启自动检测（默认开启），则根据文件扩展名检测
		if req.AutoDetect {
			bucketKey = s.cosManager.DetectBucketKeyByFileName(req.FileName)
			s.log.WithContext(ctx).Infof("自动检测存储桶: fileName=%s -> bucketKey=%s", req.FileName, bucketKey)
		}
		// 如果仍为空，将使用默认存储桶（在 COSManager 中处理）
	}

	// 构建上传选项
	opts := &pkg.UploadOptions{
		FileName:    req.FileName,
		ContentType: req.ContentType,
		BucketKey:   bucketKey,
		FileSize:    req.FileSize,
	}

	// 获取预签名 URL
	result, err := s.cosManager.GetUploadPresignedURL(ctx, opts)
	if err != nil {
		s.log.WithContext(ctx).Errorf("生成预签名 URL 失败: %v", err)
		return nil, v1.ErrorSystemError("生成上传链接失败: %s", err.Error())
	}

	return &v1.GetUploadPresignedUrlReply{
		UploadUrl:  result.UploadURL,
		FileKey:    result.FileKey,
		AccessUrl:  result.AccessURL,
		ExpireTime: result.ExpireTime,
		BucketKey:  result.BucketKey,
		BucketName: result.BucketName,
		Region:     result.Region,
	}, nil
}

// ListBuckets 获取可用的存储桶列表
func (s *FileService) ListBuckets(ctx context.Context, req *v1.ListBucketsRequest) (*v1.ListBucketsReply, error) {
	s.log.WithContext(ctx).Info("获取存储桶列表")

	// 检查 COS Manager 是否可用
	if s.cosManager == nil {
		s.log.WithContext(ctx).Error("COS Manager 未初始化")
		return nil, v1.ErrorSystemError("文件上传服务暂不可用，请联系管理员配置 COS")
	}

	// 获取所有存储桶配置
	bucketKeys := s.cosManager.GetBucketKeys()
	buckets := make([]*v1.BucketInfo, 0, len(bucketKeys))

	for _, key := range bucketKeys {
		config, ok := s.cosManager.GetBucketConfig(key)
		if !ok {
			continue
		}
		buckets = append(buckets, &v1.BucketInfo{
			Key:               key,
			BucketName:        config.Name,
			Region:            config.Region,
			AllowedExtensions: config.AllowedExtensions,
			MaxSize:           config.MaxSize,
		})
	}

	return &v1.ListBucketsReply{
		Buckets:       buckets,
		DefaultBucket: s.cosManager.GetDefaultBucketKey(),
	}, nil
}
