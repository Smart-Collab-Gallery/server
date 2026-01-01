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

// COSManager 腾讯云 COS 管理器
type COSManager struct {
	secretID          string
	secretKey         string
	defaultBucketURL  string
	defaultRegion     string
	defaultBucketName string
	defaultUploadDir  string
	log               *log.Helper
}

// NewCOSManager 创建 COS 管理器
func NewCOSManager(c *conf.Cos, logger log.Logger) (*COSManager, error) {
	helper := log.NewHelper(logger)

	// 如果配置为空，返回 nil（允许 COS 功能可选）
	if c == nil {
		helper.Warn("COS 配置为空，COS 功能将不可用")
		return nil, nil
	}

	// 检查必要的配置字段（至少需要 SecretId 和 SecretKey）
	if c.SecretId == "" || c.SecretKey == "" {
		helper.Warn("COS 配置不完整（缺少 SecretId 或 SecretKey），COS 功能将不可用")
		return nil, nil
	}

	manager := &COSManager{
		secretID:          c.SecretId,
		secretKey:         c.SecretKey,
		defaultBucketURL:  c.DefaultBucketUrl,
		defaultRegion:     c.DefaultRegion,
		defaultBucketName: c.DefaultBucketName,
		defaultUploadDir:  c.DefaultUploadDir,
		log:               helper,
	}

	helper.Info("COS Manager 初始化成功（支持动态 Bucket）")
	return manager, nil
}

// UploadOptions 上传选项
type UploadOptions struct {
	FileName    string // 原始文件名（必填）
	ContentType string // 文件 MIME 类型（可选）
	BucketName  string // 存储桶名称（可选，不传使用默认配置）
	Region      string // 地域（可选，不传使用默认配置）
	UploadDir   string // 上传目录前缀（可选，不传使用默认配置）
}

// GetUploadPresignedURL 获取上传预签名 URL（支持动态 Bucket）
func (m *COSManager) GetUploadPresignedURL(ctx context.Context, opts *UploadOptions) (*PresignedURLResult, error) {
	// 确定使用的 Bucket 名称
	bucketName := opts.BucketName
	if bucketName == "" {
		bucketName = m.defaultBucketName
	}
	if bucketName == "" {
		m.log.Error("BucketName 未指定且无默认配置")
		return nil, fmt.Errorf("bucket name is required")
	}

	// 确定使用的 Region
	region := opts.Region
	if region == "" {
		region = m.defaultRegion
	}
	if region == "" {
		m.log.Error("Region 未指定且无默认配置")
		return nil, fmt.Errorf("region is required")
	}

	// 确定使用的上传目录
	uploadDir := opts.UploadDir
	if uploadDir == "" {
		uploadDir = m.defaultUploadDir
	}

	// 构建 Bucket URL
	bucketURL := fmt.Sprintf("https://%s.cos.%s.myqcloud.com", bucketName, region)

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
	fileKey := generateFileKey(opts.FileName, uploadDir)

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

	// 构建访问 URL（上传成功后可以通过这个 URL 访问文件）
	accessURL := fmt.Sprintf("%s/%s", bucketURL, fileKey)

	result := &PresignedURLResult{
		UploadURL:  presignedURL.String(),
		FileKey:    fileKey,
		AccessURL:  accessURL,
		ExpireTime: expireTime,
		BucketName: bucketName,
		Region:     region,
	}

	m.log.Infof("生成预签名 URL 成功: bucket=%s, region=%s, fileKey=%s, expire=%s",
		bucketName, region, fileKey, time.Unix(expireTime, 0).Format(time.RFC3339))
	return result, nil
}

// generateFileKey 生成文件存储路径（key）
// 格式：{uploadDir}/YYYY/MM/DD/uuid_原始文件名
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
	BucketName string // 实际使用的存储桶名称
	Region     string // 实际使用的地域
}
