package app

import (
	"context"
	"time"

	"light-tracking/internal/models"
)

// App struct holds the application state
type App struct {
	ctx                context.Context
	database           *Database
	timer              *Timer
	systrayManager     *SystrayManager
	notificationManager *NotificationManager
}

// NewApp creates a new App application struct
func NewApp() (*App, error) {
	db, err := NewDatabase()
	if err != nil {
		return nil, err
	}

	app := &App{
		database:           db,
		timer:              NewTimer(),
		systrayManager:     nil, // Will be set in Startup
		notificationManager: nil, // Will be set in Startup
	}

	// Load active slot from database on startup
	if err := app.timer.LoadActiveSlot(db); err != nil {
		return nil, err
	}

	return app, nil
}

// Startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	// Initialize systray with delay to let Wails/GTK fully initialize
	go func() {
		time.Sleep(500 * time.Millisecond) // Wait for Wails/GTK to fully initialize
		a.systrayManager = NewSystrayManager(a)
		a.systrayManager.Run(ctx)
	}()
	// Initialize notifications
	a.notificationManager = NewNotificationManager(a)
	a.notificationManager.Start(ctx)
}

// StartTimer starts tracking time for a task
func (a *App) StartTimer(taskName string) (*models.TimeSlot, error) {
	if taskName == "" {
		return nil, nil
	}
	return a.timer.Start(taskName, a.database)
}

// StopTimer stops the current timer
func (a *App) StopTimer() (*models.TimeSlot, error) {
	return a.timer.Stop(a.database)
}

// GetActiveTimeSlot returns the currently active time slot
func (a *App) GetActiveTimeSlot() *models.TimeSlot {
	return a.timer.GetActiveSlot()
}

// IsTimerRunning returns whether the timer is currently running
func (a *App) IsTimerRunning() bool {
	return a.timer.IsRunning()
}

// GetElapsedTime returns the elapsed time for the current session
func (a *App) GetElapsedTime() int64 {
	return int64(a.timer.GetElapsedTime().Seconds())
}

// GetTimeSlotsByDate returns all time slots for a specific date
// date should be in format "2006-01-02" (YYYY-MM-DD)
func (a *App) GetTimeSlotsByDate(dateStr string) ([]*models.TimeSlot, error) {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, err
	}
	return a.database.GetTimeSlotsByDate(date)
}

// GetTaskStatistics returns aggregated statistics by task name for a specific date
// date should be in format "2006-01-02" (YYYY-MM-DD)
func (a *App) GetTaskStatistics(dateStr string) (map[string]int64, error) {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, err
	}
	return a.database.GetTaskStatistics(date)
}

// UpdateTimeSlot updates a time slot
// startTime and endTime should be in RFC3339 format (ISO 8601)
// endTime can be empty string for active slots
func (a *App) UpdateTimeSlot(id int64, taskName string, startTimeStr string, endTimeStr string) error {
	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		return err
	}

	var endTime *time.Time
	if endTimeStr != "" {
		et, err := time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			return err
		}
		endTime = &et
	}

	return a.database.UpdateTimeSlot(id, taskName, startTime, endTime)
}

// DeleteTimeSlot deletes a time slot
func (a *App) DeleteTimeSlot(id int64) error {
	return a.database.DeleteTimeSlot(id)
}

// Close closes the database connection
func (a *App) Close() error {
	return a.database.Close()
}

