package data

import (
	"context"
	"smart-collab-gallery-server/internal/conf"
	"smart-collab-gallery-server/internal/pkg"

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
	db          *gorm.DB
	rdb         *redis.Client
	emailSender *pkg.EmailSender
}

// NewData .
func NewData(c *conf.Bootstrap, logger log.Logger) (*Data, func(), error) {
	log := log.NewHelper(logger)

	// 初始化数据库连接
	db, err := gorm.Open(mysql.Open(c.Data.Database.Source), &gorm.Config{})
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
		Network:      c.Data.Redis.Network,
		Addr:         c.Data.Redis.Addr,
		Password:     c.Data.Redis.Password,
		DB:           int(c.Data.Redis.Db),
		ReadTimeout:  c.Data.Redis.ReadTimeout.AsDuration(),
		WriteTimeout: c.Data.Redis.WriteTimeout.AsDuration(),
	})

	// 测试 Redis 连接
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Errorf("failed to connect redis: %v", err)
		return nil, nil, err
	}

	// 初始化邮件发送器
	emailSender := pkg.NewEmailSender(&pkg.EmailConfig{
		SMTPHost:     c.Email.SmtpHost,
		SMTPPort:     int(c.Email.SmtpPort),
		SMTPUser:     c.Email.SmtpUser,
		SMTPPassword: c.Email.SmtpPassword,
		FromEmail:    c.Email.FromEmail,
		FromName:     c.Email.FromName,
	})

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

	return &Data{db: db, rdb: rdb, emailSender: emailSender}, cleanup, nil
}
