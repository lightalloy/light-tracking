package app

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"light-tracking/internal/models"

	_ "modernc.org/sqlite"
)

type Database struct {
	db *sql.DB
}

// NewDatabase creates a new database connection
func NewDatabase() (*Database, error) {
	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	// Create app data directory
	appDataDir := filepath.Join(homeDir, ".light-tracking")
	if err := os.MkdirAll(appDataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create app data directory: %w", err)
	}

	// Database file path
	dbPath := filepath.Join(appDataDir, "time_tracking.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	database := &Database{db: db}

	// Initialize schema
	if err := database.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return database, nil
}

// initSchema creates the database tables
func (d *Database) initSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS time_slots (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task_name TEXT NOT NULL,
		start_time DATETIME NOT NULL,
		end_time DATETIME,
		duration_seconds INTEGER DEFAULT 0
	);
	
	CREATE INDEX IF NOT EXISTS idx_start_time ON time_slots(start_time);
	CREATE INDEX IF NOT EXISTS idx_task_name ON time_slots(task_name);
	`

	_, err := d.db.Exec(query)
	return err
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// CreateTimeSlot creates a new time slot
func (d *Database) CreateTimeSlot(taskName string, startTime time.Time) (*models.TimeSlot, error) {
	query := `INSERT INTO time_slots (task_name, start_time) VALUES (?, ?)`
	result, err := d.db.Exec(query, taskName, startTime)
	if err != nil {
		return nil, fmt.Errorf("failed to create time slot: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return &models.TimeSlot{
		ID:        id,
		TaskName:  taskName,
		StartTime: startTime,
	}, nil
}

// GetActiveTimeSlot returns the currently active time slot, if any
func (d *Database) GetActiveTimeSlot() (*models.TimeSlot, error) {
	query := `SELECT id, task_name, start_time, end_time, duration_seconds 
	          FROM time_slots 
	          WHERE end_time IS NULL 
	          ORDER BY start_time DESC 
	          LIMIT 1`

	var ts models.TimeSlot
	var endTime sql.NullTime

	err := d.db.QueryRow(query).Scan(
		&ts.ID,
		&ts.TaskName,
		&ts.StartTime,
		&endTime,
		&ts.DurationSeconds,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get active time slot: %w", err)
	}

	if endTime.Valid {
		ts.EndTime = &endTime.Time
	}

	return &ts, nil
}

// StopTimeSlot stops an active time slot
func (d *Database) StopTimeSlot(id int64, endTime time.Time) error {
	// First get the start time
	var startTime time.Time
	err := d.db.QueryRow("SELECT start_time FROM time_slots WHERE id = ?", id).Scan(&startTime)
	if err != nil {
		return fmt.Errorf("failed to get start time: %w", err)
	}

	// Calculate duration
	durationSeconds := int64(endTime.Sub(startTime).Seconds())

	// Update the time slot
	query := `UPDATE time_slots 
	          SET end_time = ?, duration_seconds = ?
	          WHERE id = ?`
	
	_, err = d.db.Exec(query, endTime, durationSeconds, id)
	if err != nil {
		return fmt.Errorf("failed to stop time slot: %w", err)
	}

	return nil
}

// GetTimeSlotsByDate returns all time slots for a specific date
func (d *Database) GetTimeSlotsByDate(date time.Time) ([]*models.TimeSlot, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	query := `SELECT id, task_name, start_time, end_time, duration_seconds 
	          FROM time_slots 
	          WHERE start_time >= ? AND start_time < ?
	          ORDER BY start_time ASC`

	rows, err := d.db.Query(query, startOfDay, endOfDay)
	if err != nil {
		return nil, fmt.Errorf("failed to query time slots: %w", err)
	}
	defer rows.Close()

	var slots []*models.TimeSlot
	for rows.Next() {
		var ts models.TimeSlot
		var endTime sql.NullTime

		err := rows.Scan(
			&ts.ID,
			&ts.TaskName,
			&ts.StartTime,
			&endTime,
			&ts.DurationSeconds,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan time slot: %w", err)
		}

		if endTime.Valid {
			ts.EndTime = &endTime.Time
		}

		slots = append(slots, &ts)
	}

	return slots, rows.Err()
}

// GetTaskStatistics returns aggregated statistics by task name for a specific date
func (d *Database) GetTaskStatistics(date time.Time) (map[string]int64, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	query := `SELECT task_name, SUM(duration_seconds) as total_seconds
	          FROM time_slots 
	          WHERE start_time >= ? AND start_time < ? AND end_time IS NOT NULL
	          GROUP BY task_name
	          ORDER BY total_seconds DESC`

	rows, err := d.db.Query(query, startOfDay, endOfDay)
	if err != nil {
		return nil, fmt.Errorf("failed to query task statistics: %w", err)
	}
	defer rows.Close()

	stats := make(map[string]int64)
	for rows.Next() {
		var taskName string
		var totalSeconds int64

		err := rows.Scan(&taskName, &totalSeconds)
		if err != nil {
			return nil, fmt.Errorf("failed to scan statistics: %w", err)
		}

		stats[taskName] = totalSeconds
	}

	return stats, rows.Err()
}

// UpdateTimeSlot updates a time slot
func (d *Database) UpdateTimeSlot(id int64, taskName string, startTime time.Time, endTime *time.Time) error {
	var durationSeconds int64
	if endTime != nil {
		durationSeconds = int64(endTime.Sub(startTime).Seconds())
	}

	query := `UPDATE time_slots 
	          SET task_name = ?, start_time = ?, end_time = ?, duration_seconds = ?
	          WHERE id = ?`

	_, err := d.db.Exec(query, taskName, startTime, endTime, durationSeconds, id)
	if err != nil {
		return fmt.Errorf("failed to update time slot: %w", err)
	}

	return nil
}

// DeleteTimeSlot deletes a time slot
func (d *Database) DeleteTimeSlot(id int64) error {
	query := `DELETE FROM time_slots WHERE id = ?`
	_, err := d.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete time slot: %w", err)
	}
	return nil
}

// GetAllTimeSlots returns all time slots (for debugging/admin purposes)
func (d *Database) GetAllTimeSlots() ([]*models.TimeSlot, error) {
	query := `SELECT id, task_name, start_time, end_time, duration_seconds 
	          FROM time_slots 
	          ORDER BY start_time DESC`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all time slots: %w", err)
	}
	defer rows.Close()

	var slots []*models.TimeSlot
	for rows.Next() {
		var ts models.TimeSlot
		var endTime sql.NullTime

		err := rows.Scan(
			&ts.ID,
			&ts.TaskName,
			&ts.StartTime,
			&endTime,
			&ts.DurationSeconds,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan time slot: %w", err)
		}

		if endTime.Valid {
			ts.EndTime = &endTime.Time
		}

		slots = append(slots, &ts)
	}

	return slots, rows.Err()
}

