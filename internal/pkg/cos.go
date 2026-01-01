package pkg

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"smart-collab-gallery-server/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"github.com/tencentyun/cos-go-sdk-v5"
)

// BucketConfig 单个存储桶配置
type BucketConfig struct {
	Name              string   // 存储桶名称
	Region            string   // 地域
	UploadDir         string   // 上传目录前缀
	AllowedExtensions []string // 允许的文件扩展名
	MaxSize           int64    // 最大文件大小（字节）
}

// COSManager 腾讯云 COS 管理器（支持多存储桶）
type COSManager struct {
	secretID      string
	secretKey     string
	buckets       map[string]*BucketConfig // 多存储桶配置
	defaultBucket string                   // 默认存储桶 key
	log           *log.Helper
}

// NewCOSManager 创建 COS 管理器
func NewCOSManager(c *conf.Cos, logger log.Logger) (*COSManager, error) {
	helper := log.NewHelper(logger)

	// 如果配置为空，返回 nil（允许 COS 功能可选）
	if c == nil {
		helper.Warn("COS 配置为空，COS 功能将不可用")
		return nil, nil
	}

	// 检查必要的配置字段
	if c.SecretId == "" || c.SecretKey == "" {
		helper.Warn("COS 配置不完整（缺少 SecretId 或 SecretKey），COS 功能将不可用")
		return nil, nil
	}

	// 检查存储桶配置
	if len(c.Buckets) == 0 {
		helper.Warn("COS 未配置任何存储桶，COS 功能将不可用")
		return nil, nil
	}

	// 转换存储桶配置
	buckets := make(map[string]*BucketConfig, len(c.Buckets))
	for key, bucket := range c.Buckets {
		buckets[key] = &BucketConfig{
			Name:              bucket.BucketName,
			Region:            bucket.Region,
			UploadDir:         bucket.UploadDir,
			AllowedExtensions: bucket.AllowedExtensions,
			MaxSize:           bucket.MaxSize,
		}
		helper.Infof("加载存储桶配置: key=%s, bucket=%s, region=%s", key, bucket.BucketName, bucket.Region)
	}

	// 验证默认存储桶
	defaultBucket := c.DefaultBucket
	if defaultBucket == "" {
		// 如果未指定默认存储桶，使用第一个
		for key := range buckets {
			defaultBucket = key
			break
		}
	}
	if _, ok := buckets[defaultBucket]; !ok {
		helper.Warnf("默认存储桶 '%s' 不存在，将使用第一个存储桶", defaultBucket)
		for key := range buckets {
			defaultBucket = key
			break
		}
	}

	manager := &COSManager{
		secretID:      c.SecretId,
		secretKey:     c.SecretKey,
		buckets:       buckets,
		defaultBucket: defaultBucket,
		log:           helper,
	}

	helper.Infof("COS Manager 初始化成功，共 %d 个存储桶，默认: %s", len(buckets), defaultBucket)
	return manager, nil
}

// GetBucketKeys 获取所有存储桶的 key 列表
func (m *COSManager) GetBucketKeys() []string {
	keys := make([]string, 0, len(m.buckets))
	for key := range m.buckets {
		keys = append(keys, key)
	}
	return keys
}

// GetBucketConfig 获取指定存储桶配置
func (m *COSManager) GetBucketConfig(bucketKey string) (*BucketConfig, bool) {
	config, ok := m.buckets[bucketKey]
	return config, ok
}

// GetDefaultBucketKey 获取默认存储桶 key
func (m *COSManager) GetDefaultBucketKey() string {
	return m.defaultBucket
}

// UploadOptions 上传选项
type UploadOptions struct {
	FileName    string // 原始文件名（必填）
	ContentType string // 文件 MIME 类型（可选）
	BucketKey   string // 存储桶 key（如 image/video/document，可选）
	FileSize    int64  // 文件大小（字节，用于校验，可选）
}

// GetUploadPresignedURL 获取上传预签名 URL
func (m *COSManager) GetUploadPresignedURL(ctx context.Context, opts *UploadOptions) (*PresignedURLResult, error) {
	// 确定使用的存储桶
	bucketKey := opts.BucketKey
	if bucketKey == "" {
		bucketKey = m.defaultBucket
	}

	bucketConfig, ok := m.buckets[bucketKey]
	if !ok {
		m.log.Errorf("存储桶 '%s' 不存在", bucketKey)
		return nil, fmt.Errorf("bucket '%s' not found", bucketKey)
	}

	// 校验文件扩展名
	if len(bucketConfig.AllowedExtensions) > 0 {
		ext := strings.ToLower(path.Ext(opts.FileName))
		allowed := false
		for _, allowedExt := range bucketConfig.AllowedExtensions {
			if ext == strings.ToLower(allowedExt) {
				allowed = true
				break
			}
		}
		if !allowed {
			m.log.Warnf("文件扩展名 '%s' 不被存储桶 '%s' 允许", ext, bucketKey)
			return nil, fmt.Errorf("file extension '%s' not allowed for bucket '%s'", ext, bucketKey)
		}
	}

	// 校验文件大小
	if bucketConfig.MaxSize > 0 && opts.FileSize > bucketConfig.MaxSize {
		m.log.Warnf("文件大小 %d 超过存储桶 '%s' 限制 %d", opts.FileSize, bucketKey, bucketConfig.MaxSize)
		return nil, fmt.Errorf("file size %d exceeds limit %d for bucket '%s'", opts.FileSize, bucketConfig.MaxSize, bucketKey)
	}

	// 构建 Bucket URL
	bucketURL := fmt.Sprintf("https://%s.cos.%s.myqcloud.com", bucketConfig.Name, bucketConfig.Region)

	// 解析 Bucket URL
	u, err := url.Parse(bucketURL)
	if err != nil {
		m.log.Errorf("解析 Bucket URL 失败: %v", err)
		return nil, fmt.Errorf("invalid bucket url: %w", err)
	}

	// 为当前请求创建临时 COS 客户端
	baseURL := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(baseURL, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  m.secretID,
			SecretKey: m.secretKey,
		},
	})

	// 生成唯一的文件 key（路径）
	fileKey := generateFileKey(opts.FileName, bucketConfig.UploadDir)

	// 设置过期时间为 10 分钟
	expireDuration := 10 * time.Minute

	// 准备签名选项
	signOpt := &cos.PresignedURLOptions{
		Query:  &url.Values{},
		Header: &http.Header{},
	}

	// 如果提供了 Content-Type，添加到签名中
	if opts.ContentType != "" {
		signOpt.Header.Set("Content-Type", opts.ContentType)
	}

	// 获取预签名 URL
	presignedURL, err := client.Object.GetPresignedURL(
		ctx,
		http.MethodPut,
		fileKey,
		m.secretID,
		m.secretKey,
		expireDuration,
		signOpt,
	)
	if err != nil {
		m.log.Errorf("生成预签名 URL 失败: %v", err)
		return nil, fmt.Errorf("failed to generate presigned url: %w", err)
	}

	// 计算过期时间戳
	expireTime := time.Now().Add(expireDuration).Unix()

	// 构建访问 URL
	accessURL := fmt.Sprintf("%s/%s", bucketURL, fileKey)

	result := &PresignedURLResult{
		UploadURL:  presignedURL.String(),
		FileKey:    fileKey,
		AccessURL:  accessURL,
		ExpireTime: expireTime,
		BucketKey:  bucketKey,
		BucketName: bucketConfig.Name,
		Region:     bucketConfig.Region,
	}

	m.log.Infof("生成预签名 URL 成功: bucketKey=%s, bucket=%s, region=%s, fileKey=%s",
		bucketKey, bucketConfig.Name, bucketConfig.Region, fileKey)
	return result, nil
}

// DetectBucketKeyByFileName 根据文件名自动检测应使用的存储桶 key
func (m *COSManager) DetectBucketKeyByFileName(fileName string) string {
	ext := strings.ToLower(path.Ext(fileName))

	for key, config := range m.buckets {
		for _, allowedExt := range config.AllowedExtensions {
			if ext == strings.ToLower(allowedExt) {
				return key
			}
		}
	}

	// 未匹配到，返回默认存储桶
	return m.defaultBucket
}

// generateFileKey 生成文件存储路径（key）
func generateFileKey(fileName, uploadDir string) string {
	now := time.Now()

	// 处理上传目录前缀
	dirPrefix := uploadDir
	if dirPrefix != "" && !strings.HasSuffix(dirPrefix, "/") {
		dirPrefix += "/"
	}

	datePrefix := fmt.Sprintf("%s%04d/%02d/%02d/",
		dirPrefix,
		now.Year(),
		now.Month(),
		now.Day(),
	)

	// 生成 UUID 作为文件前缀
	uniqueID := uuid.New().String()

	// 获取文件扩展名
	ext := path.Ext(fileName)
	baseName := fileName[:len(fileName)-len(ext)]

	// 清理文件名
	cleanName := cleanFileName(baseName)

	// 组合最终的文件 key
	fileKey := fmt.Sprintf("%s%s_%s%s", datePrefix, uniqueID, cleanName, ext)

	return fileKey
}

// cleanFileName 清理文件名中的特殊字符
func cleanFileName(name string) string {
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
	BucketKey  string // 存储桶 key（如 image/video/document）
	BucketName string // 实际使用的存储桶名称
	Region     string // 实际使用的地域
}
