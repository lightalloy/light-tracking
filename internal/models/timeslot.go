package models

import "time"

// TimeSlot represents a time tracking entry
type TimeSlot struct {
	ID              int64     `json:"id"`
	TaskName        string    `json:"task_name"`
	StartTime       time.Time `json:"start_time"`
	EndTime         *time.Time `json:"end_time,omitempty"`
	DurationSeconds int64     `json:"duration_seconds"`
}

// IsActive returns true if the time slot is currently active (no end time)
func (ts *TimeSlot) IsActive() bool {
	return ts.EndTime == nil
}

// CalculateDuration calculates and sets the duration in seconds
func (ts *TimeSlot) CalculateDuration() {
	if ts.EndTime != nil {
		ts.DurationSeconds = int64(ts.EndTime.Sub(ts.StartTime).Seconds())
	}
}

