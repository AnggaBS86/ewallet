package db

import (
	"time"

	"ewallet/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func New(cfg config.Config) (*gorm.DB, error) {
	gormCfg := &gorm.Config{}
	gormCfg.Logger = logger.Default.LogMode(logger.Warn)

	db, err := gorm.Open(postgres.Open(cfg.DBDSN), gormCfg)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// this is for connection pool setup
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	return db, nil
}
