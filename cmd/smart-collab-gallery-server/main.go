package main

import (
	"flag"
	"os"
	"path/filepath"

	"smart-collab-gallery-server/internal/conf"
	"smart-collab-gallery-server/internal/pkg"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	_ "go.uber.org/automaxprocs"
)

var (
	// Name is the name of the compiled software.
	Name string
	// Version is the version of the compiled software.
	Version string
	// ConsulAddress Consul 服务地址
	ConsulAddress string
	// ConsulToken Consul 访问令牌
	ConsulToken string
	// flagconf is the config flag.
	flagconf string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")

	// 从环境变量读取配置
	if name := os.Getenv("APP_NAME"); name != "" {
		Name = name
	}
	if version := os.Getenv("APP_VERSION"); version != "" {
		Version = version
	}
	if addr := os.Getenv("CONSUL_ADDRESS"); addr != "" {
		ConsulAddress = addr
	}
	if token := os.Getenv("CONSUL_TOKEN"); token != "" {
		ConsulToken = token
	}
}

func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			gs,
			hs,
		),
	)
}

func main() {
	flag.Parse()
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)

	configPath := filepath.Join(flagconf, "config.yaml")

	// 1. 如果 Name 和 ConsulAddress 都不为空，尝试从 Consul 加载配置
	// 使用 Name 作为 Consul Key
	if Name != "" && ConsulAddress != "" {
		log.NewHelper(logger).Infof("检测到 Consul 环境变量，尝试从 Consul 加载配置")

		// 使用环境变量中的 Consul 配置
		consulConfig := pkg.ConsulConfig{
			Enabled: true,
			Address: ConsulAddress,
			Token:   ConsulToken,
		}
		consulLoader := pkg.NewConsulConfigLoader(consulConfig, logger)
		if err := consulLoader.LoadAndWriteConfig(configPath, Name); err != nil {
			log.NewHelper(logger).Warnf("从 Consul 读取配置失败，使用本地配置文件: %v", err)
		}
	} else if Name != "" {
		// 如果只有 Name，尝试从本地配置文件读取 Consul 设置
		tmpConfig := config.New(
			config.WithSource(
				file.NewSource(configPath),
			),
		)
		if err := tmpConfig.Load(); err == nil {
			var tmpBc conf.Bootstrap
			if err := tmpConfig.Scan(&tmpBc); err == nil {
				if tmpBc.Consul != nil && tmpBc.Consul.Enabled {
					// 使用配置文件中的 Consul 设置
					consulConfig := pkg.ConsulConfig{
						Enabled: true,
						Address: tmpBc.Consul.Address,
						Token:   "",
					}
					consulLoader := pkg.NewConsulConfigLoader(consulConfig, logger)
					if err := consulLoader.LoadAndWriteConfig(configPath, Name); err != nil {
						log.NewHelper(logger).Warnf("从 Consul 读取配置失败，使用本地配置文件: %v", err)
					}
				}
			}
		}
		tmpConfig.Close()
	}

	// 2. 加载本地配置文件（无论 Consul 是否成功）
	// 只加载 config.yaml，不加载其他文件（如 config.yaml.example）
	c := config.New(
		config.WithSource(
			file.NewSource(configPath),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	app, cleanup, err := wireApp(bc.Server, bc.Data, bc.Auth, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}
