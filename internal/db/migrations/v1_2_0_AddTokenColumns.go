package db

import (
	"gpt-load/internal/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// V1_2_0_AddTokenColumns ensures that token-related columns exist in the request_logs table.
func V1_2_0_AddTokenColumns(db *gorm.DB) error {
	// GORM AutoMigrate should handle this, but we've seen cases where it doesn't 
	// add new columns to existing SQLite tables if they were added later.
	
	fields := []string{"prompt_tokens", "completion_tokens", "total_tokens"}
	
	for _, field := range fields {
		if !db.Migrator().HasColumn(&models.RequestLog{}, field) {
			logrus.Infof("Adding missing column %s to request_logs table", field)
			if err := db.Migrator().AddColumn(&models.RequestLog{}, field); err != nil {
				logrus.Errorf("Failed to add column %s: %v", field, err)
				// Continue with other fields
			}
		}
	}

	return nil
}
