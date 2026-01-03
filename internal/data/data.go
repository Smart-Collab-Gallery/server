package data

import (
	"context"
	"smart-collab-gallery-server/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewGreeterRepo, NewUserRepo)

// Data .
type Data struct {
	db  *gorm.DB
	rdb *redis.Client
}

// NewData .
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	log := log.NewHelper(logger)

	// 初始化数据库连接
	db, err := gorm.Open(mysql.Open(c.Database.Source), &gorm.Config{})
	if err != nil {
		log.Errorf("failed to connect database: %v", err)
		return nil, nil, err
	}

	// 自动迁移数据表
	if err := db.AutoMigrate(&User{}); err != nil {
		log.Errorf("failed to migrate database: %v", err)
		return nil, nil, err
	}

	// 初始化 Redis 连接
	rdb := redis.NewClient(&redis.Options{
		Network:      c.Redis.Network,
		Addr:         c.Redis.Addr,
		Password:     c.Redis.Password,
		DB:           int(c.Redis.Db),
		ReadTimeout:  c.Redis.ReadTimeout.AsDuration(),
		WriteTimeout: c.Redis.WriteTimeout.AsDuration(),
	})

	// 测试 Redis 连接
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Errorf("failed to connect redis: %v", err)
		return nil, nil, err
	}

	cleanup := func() {
		log.Info("closing the data resources")
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
		if rdb != nil {
			rdb.Close()
		}
	}

	return &Data{db: db, rdb: rdb}, cleanup, nil
}
