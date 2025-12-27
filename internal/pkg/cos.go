package pkg

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"

	"smart-collab-gallery-server/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"github.com/tencentyun/cos-go-sdk-v5"
)

// COSManager 腾讯云 COS 管理器
type COSManager struct {
	client     *cos.Client
	secretID   string
	secretKey  string
	bucketURL  string
	region     string
	bucketName string
	uploadDir  string
	log        *log.Helper
}

// NewCOSManager 创建 COS 管理器
func NewCOSManager(c *conf.Cos, logger log.Logger) (*COSManager, error) {
	helper := log.NewHelper(logger)

	// 解析 Bucket URL
	u, err := url.Parse(c.BucketUrl)
	if err != nil {
		helper.Errorf("解析 Bucket URL 失败: %v", err)
		return nil, fmt.Errorf("invalid bucket url: %w", err)
	}

	// 创建 COS 客户端
	baseURL := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(baseURL, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  c.SecretId,
			SecretKey: c.SecretKey,
		},
	})

	manager := &COSManager{
		client:     client,
		secretID:   c.SecretId,
		secretKey:  c.SecretKey,
		bucketURL:  c.BucketUrl,
		region:     c.Region,
		bucketName: c.BucketName,
		uploadDir:  c.UploadDir,
		log:        helper,
	}

	helper.Info("COS Manager 初始化成功")
	return manager, nil
}

// GetUploadPresignedURL 获取上传预签名 URL
// fileName: 原始文件名
// contentType: 文件 MIME 类型（可选）
// expire: 过期时间（默认 10 分钟）
func (m *COSManager) GetUploadPresignedURL(ctx context.Context, fileName string, contentType string) (*PresignedURLResult, error) {
	// 生成唯一的文件 key（路径）
	fileKey := m.generateFileKey(fileName)

	// 设置过期时间为 10 分钟
	expireDuration := 10 * time.Minute

	// 准备签名选项
	opt := &cos.PresignedURLOptions{
		Query:  &url.Values{},
		Header: &http.Header{},
	}

	// 如果提供了 Content-Type，添加到签名中
	if contentType != "" {
		opt.Header.Set("Content-Type", contentType)
	}

	// 获取预签名 URL
	presignedURL, err := m.client.Object.GetPresignedURL(
		ctx,
		http.MethodPut,
		fileKey,
		m.secretID,
		m.secretKey,
		expireDuration,
		opt,
	)
	if err != nil {
		m.log.Errorf("生成预签名 URL 失败: %v", err)
		return nil, fmt.Errorf("failed to generate presigned url: %w", err)
	}

	// 计算过期时间戳
	expireTime := time.Now().Add(expireDuration).Unix()

	// 构建访问 URL（上传成功后可以通过这个 URL 访问文件）
	accessURL := fmt.Sprintf("%s/%s", m.bucketURL, fileKey)

	result := &PresignedURLResult{
		UploadURL:  presignedURL.String(),
		FileKey:    fileKey,
		AccessURL:  accessURL,
		ExpireTime: expireTime,
	}

	m.log.Infof("生成预签名 URL 成功: fileKey=%s, expire=%s", fileKey, time.Unix(expireTime, 0).Format(time.RFC3339))
	return result, nil
}

// generateFileKey 生成文件存储路径（key）
// 格式：uploads/YYYY/MM/DD/uuid_原始文件名
func (m *COSManager) generateFileKey(fileName string) string {
	now := time.Now()
	datePrefix := fmt.Sprintf("%s%04d/%02d/%02d/",
		m.uploadDir,
		now.Year(),
		now.Month(),
		now.Day(),
	)

	// 生成 UUID 作为文件前缀，避免文件名冲突
	uniqueID := uuid.New().String()

	// 获取文件扩展名
	ext := path.Ext(fileName)
	baseName := fileName[:len(fileName)-len(ext)]

	// 清理文件名（移除特殊字符）
	cleanName := cleanFileName(baseName)

	// 组合最终的文件 key
	fileKey := fmt.Sprintf("%s%s_%s%s", datePrefix, uniqueID, cleanName, ext)

	return fileKey
}

// cleanFileName 清理文件名中的特殊字符
func cleanFileName(name string) string {
	// 这里可以添加更多的清理规则
	// 简单实现：移除空格等
	result := ""
	for _, r := range name {
		if (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '-' || r == '_' {
			result += string(r)
		}
	}
	if result == "" {
		result = "file"
	}
	return result
}

// PresignedURLResult 预签名 URL 结果
type PresignedURLResult struct {
	UploadURL  string // 预签名上传 URL
	FileKey    string // 文件 key（路径）
	AccessURL  string // 访问 URL
	ExpireTime int64  // 过期时间戳
}
