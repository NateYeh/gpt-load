package db

import (
	"gpt-load/internal/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// V1_5_0_AddPerformanceIndexes ensures that performance-critical indexes exist.
func V1_5_0_AddPerformanceIndexes(db *gorm.DB) error {
	indexName := "idx_logs_dashboard_stats"
	
	// Check if index exists - GORM doesn't have a direct 'HasIndex', so we use Exec with IF NOT EXISTS for safety
	// This works for SQLite, MySQL, and PostgreSQL
	logrus.Infof("Ensuring performance index %s exists on request_logs table", indexName)
	
	var err error
	if db.Dialector.Name() == "sqlite" {
		err = db.Exec("CREATE INDEX IF NOT EXISTS " + indexName + " ON request_logs (request_type, timestamp)").Error
	} else if db.Dialector.Name() == "mysql" {
		// MySQL doesn't support CREATE INDEX IF NOT EXISTS directly in older versions, 
		// but GORM has better support for MySQL indexes
		if !db.Migrator().HasIndex(&models.RequestLog{}, indexName) {
			err = db.Exec("CREATE INDEX " + indexName + " ON request_logs (request_type, timestamp)").Error
		}
	} else {
		// Postgres supports IF NOT EXISTS
		err = db.Exec("CREATE INDEX IF NOT EXISTS " + indexName + " ON request_logs (request_type, timestamp)").Error
	}

	if err != nil {
		logrus.Errorf("Failed to create index %s: %v", indexName, err)
		// We don't return error to block startup for index failures
	}

	return nil
}