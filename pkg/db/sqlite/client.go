package sqlite

import (
	"fmt"
	"os"
	"time"

	extraClausePlugin "github.com/WinterYukky/gorm-extra-clause-plugin"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gorm_logger "gorm.io/gorm/logger"

	"github.com/erwanlbp/trading-bot/pkg/log"
	"github.com/erwanlbp/trading-bot/pkg/log/zapgorm2"
)

func NewDB(logger *log.Logger, folderName string, filename string) (*gorm.DB, error) {

	if _, err := os.Stat(folderName); os.IsNotExist(err) {
		logger.Info(fmt.Sprintf("Folder %s does not exist, creating it", folderName))
		if err := os.Mkdir(folderName, os.ModePerm); err != nil {
			return nil, fmt.Errorf("failed to create folder %s : %w", folderName, err)
		}
	}

	dbFileName := folderName + "/" + filename
	dsn := fmt.Sprintf("file:%s.db", dbFileName)
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
