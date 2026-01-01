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
// 前端可以通过 bucket_name 字段传递 bucket_key（如 image/video/document）
// 如果不传，则根据文件扩展名自动检测存储桶
func (s *FileService) GetUploadPresignedUrl(ctx context.Context, req *v1.GetUploadPresignedUrlRequest) (*v1.GetUploadPresignedUrlReply, error) {
	s.log.WithContext(ctx).Infof("获取上传预签名 URL: fileName=%s, contentType=%s, bucketKey=%s",
		req.FileName, req.ContentType, req.BucketName)

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
	// 复用 BucketName 字段作为 bucket_key（如 image/video/document）
	bucketKey := req.BucketName
	if bucketKey == "" {
		// 如果未指定，则根据文件扩展名自动检测
		bucketKey = s.cosManager.DetectBucketKeyByFileName(req.FileName)
		s.log.WithContext(ctx).Infof("自动检测存储桶: fileName=%s -> bucketKey=%s", req.FileName, bucketKey)
	}

	// 构建上传选项
	opts := &pkg.UploadOptions{
		FileName:    req.FileName,
		ContentType: req.ContentType,
		BucketKey:   bucketKey,
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
		BucketName: result.BucketKey, // 返回实际使用的 bucket key
		Region:     result.Region,
	}, nil
}
