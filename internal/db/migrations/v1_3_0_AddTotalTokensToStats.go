package db

import (
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gpt-load/internal/models"
	"time"
)

// V1_3_0_AddTotalTokensToStats ensures that total_tokens column exists in the group_hourly_stats table
// and backfills historical data from request_logs.
func V1_3_0_AddTotalTokensToStats(db *gorm.DB) error {
	columnName := "total_tokens"

	// 1. Add column if not exists
	if !db.Migrator().HasColumn(&models.GroupHourlyStat{}, columnName) {
		logrus.Infof("Adding missing column %s to group_hourly_stats table", columnName)
		if err := db.Migrator().AddColumn(&models.GroupHourlyStat{}, columnName); err != nil {
			logrus.Errorf("Failed to add column %s: %v", columnName, err)
			return err
		}
	}

	// 2. Backfill data from request_logs
	// We only backfill for success requests and 'final' type
	logrus.Info("Starting backfill for total_tokens in group_hourly_stats...")

	// Query aggregate token counts from request_logs grouped by hour and group_id
	// We limit backfill to last 14 days to avoid performance issues on large log tables
	startTime := time.Now().AddDate(0, 0, -14).Truncate(24 * time.Hour)

	type result struct {
		Hour        string
		GroupID     uint
		TotalTokens int64
	}
	var results []result

	// Use different SQL for different DB drivers
	var query string
	if db.Dialector.Name() == "sqlite" {
		query = `
			SELECT (substr(timestamp, 1, 13) || ':00:00+08:00') as hour, group_id, SUM(total_tokens) as total_tokens 
			FROM request_logs 
			WHERE timestamp >= ? AND is_success = 1 AND request_type = 'final'
			GROUP BY hour, group_id
		`
	} else if db.Dialector.Name() == "mysql" {
		query = `
			SELECT DATE_FORMAT(timestamp, '%Y-%m-%d %H:00:00') as hour, group_id, SUM(total_tokens) as total_tokens 
			FROM request_logs 
			WHERE timestamp >= ? AND is_success = 1 AND request_type = 'final'
			GROUP BY hour, group_id
		`
	} else {
		// Postgres
		query = `
			SELECT date_trunc('hour', timestamp) as hour, group_id, SUM(total_tokens) as total_tokens 
			FROM request_logs 
			WHERE timestamp >= ? AND is_success = 1 AND request_type = 'final'
			GROUP BY hour, group_id
		`
	}

	if err := db.Raw(query, startTime).Scan(&results).Error; err != nil {
		logrus.Errorf("Failed to query historical token data for backfill: %v", err)
		return nil // Don't block migration for backfill errors
	}

	if len(results) > 0 {
		logrus.Infof("Found %d hourly records to backfill", len(results))
		for _, r := range results {
			// Update the stats table.
			// We use Exec instead of Model for better precision on SQLite column types during migration
			if err := db.Exec("UPDATE group_hourly_stats SET total_tokens = ? WHERE time = ? AND group_id = ?",
				r.TotalTokens, r.Hour, r.GroupID).Error; err != nil {
				logrus.Warnf("Failed to backfill stat for %v GroupID %d: %v", r.Hour, r.GroupID, err)
			}
		}
		logrus.Info("Backfill for total_tokens completed.")
	}

	return nil
}
