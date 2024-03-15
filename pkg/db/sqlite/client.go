package sqlite

import (
	"fmt"
	"time"

	extraClausePlugin "github.com/WinterYukky/gorm-extra-clause-plugin"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gorm_logger "gorm.io/gorm/logger"

	"github.com/erwanlbp/trading-bot/pkg/log"
	"github.com/erwanlbp/trading-bot/pkg/log/zapgorm2"
)

func NewDB(logger *log.Logger, filename string) (*gorm.DB, error) {
	dsn := fmt.Sprintf("file:%s.db", filename)
	theDatabase, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: zapgorm2.Logger{
			ZapLogger:                 logger.Logger,
			LogLevel:                  gorm_logger.Error,
			SlowThreshold:             1 * time.Second,
			IgnoreRecordNotFoundError: true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init db: %w", err)
	}

	if err := theDatabase.Use(extraClausePlugin.New()); err != nil {
		return nil, fmt.Errorf("failed to add extra clause plugin: %w", err)
	}

	return theDatabase, nil
}
