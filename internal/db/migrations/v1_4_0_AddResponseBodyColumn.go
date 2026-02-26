package db

import (
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gpt-load/internal/models"
)

// V1_4_0_AddResponseBodyColumn ensures that response_body column exists in the request_logs table.
func V1_4_0_AddResponseBodyColumn(db *gorm.DB) error {
	columnName := "response_body"

	if !db.Migrator().HasColumn(&models.RequestLog{}, columnName) {
		logrus.Infof("Adding missing column %s to request_logs table", columnName)
		if err := db.Migrator().AddColumn(&models.RequestLog{}, columnName); err != nil {
			logrus.Errorf("Failed to add column %s: %v", columnName, err)
			return err
		}
	}

	return nil
}
