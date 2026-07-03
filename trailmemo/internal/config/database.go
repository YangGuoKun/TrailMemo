package config

import (
	"fmt"
	"time"

	platformlogger "github.com/trailmemo/internal/platform/logger"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

func GetDB() *gorm.DB {
	return db
}

func InitDB(cfg *DatabaseConfig) error {
	var err error

	gormConfig := &gorm.Config{
		Logger: platformlogger.NewGormLogger(configuredGormLog()),
	}

	// 1. 先连接到 MySQL 服务器（不指定具体数据库）
	//    使用 /?charset=... 来连接默认数据库，而不是 /trail_memo_db
	rootDsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=%s&parseTime=True&loc=%s",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Charset, cfg.Loc)

	dbTemp, err := gorm.Open(mysql.Open(rootDsn), gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect MySQL server: %w", err)
	}

	// 2. 创建数据库（如果不存在）
	createDBSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", cfg.Name)
	if err = dbTemp.Exec(createDBSQL).Error; err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}
	platformlogger.L().Info("database_ready",
		zap.String("event", "database_ready"),
		zap.String("database", cfg.Name),
	)

	// 3. 关闭临时连接，正式连接目标数据库
	if sqlDB, err := dbTemp.DB(); err == nil {
		sqlDB.Close()
	}

	// 4. 连接目标数据库
	db, err = gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
		Logger: platformlogger.NewGormLogger(configuredGormLog()),
	})
	if err != nil {
		return fmt.Errorf("failed to connect target database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err = sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}

func configuredGormLog() platformlogger.GormConfig {
	if Get() == nil {
		return platformlogger.DefaultConfig("release").Gorm
	}
	return Get().Log.Gorm
}

func CloseDB() error {
	if db != nil {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
