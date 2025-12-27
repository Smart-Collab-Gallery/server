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

// GetUploadPresignedUrl 获取上传预签名 URL
func (s *FileService) GetUploadPresignedUrl(ctx context.Context, req *v1.GetUploadPresignedUrlRequest) (*v1.GetUploadPresignedUrlReply, error) {
	s.log.WithContext(ctx).Infof("获取上传预签名 URL: fileName=%s, contentType=%s", req.FileName, req.ContentType)

	// 检查 COS Manager 是否可用
	if s.cosManager == nil {
		s.log.WithContext(ctx).Error("COS Manager 未初始化，请检查配置")
		return nil, v1.ErrorSystemError("文件上传服务暂不可用，请联系管理员配置 COS")
	}

	// 参数校验
	if req.FileName == "" {
		return nil, v1.ErrorParamsError("文件名不能为空")
	}

	// 获取预签名 URL
	result, err := s.cosManager.GetUploadPresignedURL(ctx, req.FileName, req.ContentType)
	if err != nil {
		s.log.WithContext(ctx).Errorf("生成预签名 URL 失败: %v", err)
		return nil, v1.ErrorSystemError("生成上传链接失败")
	}

	return &v1.GetUploadPresignedUrlReply{
		UploadUrl:  result.UploadURL,
		FileKey:    result.FileKey,
		AccessUrl:  result.AccessURL,
		ExpireTime: result.ExpireTime,
	}, nil
}
