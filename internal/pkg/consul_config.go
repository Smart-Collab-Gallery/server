package pkg

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-kratos/kratos/v2/log"
	consulapi "github.com/hashicorp/consul/api"
)

// ConsulConfig Consul 配置
type ConsulConfig struct {
	Enabled bool
	Address string
	Token   string
}

// ConsulConfigLoader Consul 配置加载器
type ConsulConfigLoader struct {
	config ConsulConfig
	logger *log.Helper
}

// NewConsulConfigLoader 创建 Consul 配置加载器
func NewConsulConfigLoader(config ConsulConfig, logger log.Logger) *ConsulConfigLoader {
	return &ConsulConfigLoader{
		config: config,
		logger: log.NewHelper(logger),
	}
}

// LoadAndWriteConfig 从 Consul 加载配置并写入本地文件
func (c *ConsulConfigLoader) LoadAndWriteConfig(configPath string, key string) error {
	// 如果未启用 Consul，直接返回
	if !c.config.Enabled {
		c.logger.Info("Consul 配置中心未启用，使用本地配置文件")
		return nil
	}

	c.logger.Infof("尝试从 Consul 加载配置: address=%s, key=%s", c.config.Address, key)

	// 创建 Consul 客户端
	consulConfig := consulapi.DefaultConfig()
	consulConfig.Address = c.config.Address
	if c.config.Token != "" {
		consulConfig.Token = c.config.Token
	}

	client, err := consulapi.NewClient(consulConfig)
	if err != nil {
		c.logger.Errorf("创建 Consul 客户端失败: %v", err)
		return fmt.Errorf("创建 Consul 客户端失败: %w", err)
	}

	// 从 Consul KV 获取配置
	kv := client.KV()
	pair, _, err := kv.Get(key, nil)
	if err != nil {
		c.logger.Errorf("从 Consul 获取配置失败: %v", err)
		return fmt.Errorf("从 Consul 获取配置失败: %w", err)
	}

	if pair == nil || len(pair.Value) == 0 {
		c.logger.Warnf("Consul 中没有找到配置 Key: %s", key)
		return fmt.Errorf("Consul 中没有找到配置 Key: %s", key)
	}

	c.logger.Infof("成功从 Consul 获取配置 Key=%s，大小: %d bytes", key, len(pair.Value))

	// 确保配置目录存在
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		c.logger.Errorf("创建配置目录失败: %v", err)
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	// 写入本地配置文件
	if err := os.WriteFile(configPath, pair.Value, 0644); err != nil {
		c.logger.Errorf("写入本地配置文件失败: %v", err)
		return fmt.Errorf("写入本地配置文件失败: %w", err)
	}

	c.logger.Infof("成功将 Consul 配置写入本地文件: %s", configPath)
	return nil
}

// SyncConfigToConsul 将本地配置同步到 Consul（可选功能，用于首次配置）
func (c *ConsulConfigLoader) SyncConfigToConsul(configPath string, key string) error {
	if !c.config.Enabled {
		return fmt.Errorf("Consul 未启用")
	}

	// 读取本地配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取本地配置文件失败: %w", err)
	}

	// 创建 Consul 客户端
	consulConfig := consulapi.DefaultConfig()
	consulConfig.Address = c.config.Address
	if c.config.Token != "" {
		consulConfig.Token = c.config.Token
	}

	client, err := consulapi.NewClient(consulConfig)
	if err != nil {
		return fmt.Errorf("创建 Consul 客户端失败: %w", err)
	}

	// 写入 Consul KV
	kv := client.KV()
	pair := &consulapi.KVPair{
		Key:   key,
		Value: data,
	}

	_, err = kv.Put(pair, nil)
	if err != nil {
		return fmt.Errorf("写入 Consul 配置失败: %w", err)
	}

	c.logger.Infof("成功将配置同步到 Consul Key: %s", key)
	return nil
}
