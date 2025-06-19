package repository

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"shortlink-platform/backend/internal/model"
	"shortlink-platform/backend/pkg/logging"
)

var DB *gorm.DB

func InitDB(logger *zap.Logger, atomicLogLevel zap.AtomicLevel) {
	dsn := viper.GetString("db.dsn")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logging.NewGormLogger(logger, logging.ToGormLogLevel(atomicLogLevel.Level())), // 注入 logger 并转换级别
	})
	if err != nil {
		logging.Logger.Fatal("Failed to connect database", zap.Error(err))
	}

	err = db.AutoMigrate(&model.ShortLink{}, &model.DailyStat{}, &model.WhitelistDomain{})
	if err != nil {
		logging.Logger.Fatal("Failed to migrate database", zap.Error(err))
	}

	DB = db
}
