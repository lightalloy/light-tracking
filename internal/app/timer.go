package app

import (
	"sync"
	"time"

	"light-tracking/internal/models"
)

type Timer struct {
	mu            sync.RWMutex
	activeSlot    *models.TimeSlot
	isRunning     bool
	startTime     time.Time
	notifyChannel chan bool
}

func NewTimer() *Timer {
	return &Timer{
		notifyChannel: make(chan bool, 1),
	}
}

// Start starts the timer with a task name
func (t *Timer) Start(taskName string, db *Database) (*models.TimeSlot, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// If there's an active slot, stop it first
	if t.activeSlot != nil && t.activeSlot.IsActive() {
		err := db.StopTimeSlot(t.activeSlot.ID, time.Now())
		if err != nil {
			return nil, err
		}
	}

	// Create new time slot
	now := time.Now()
	slot, err := db.CreateTimeSlot(taskName, now)
	if err != nil {
		return nil, err
	}

	t.activeSlot = slot
	t.isRunning = true
	t.startTime = now

	// Notify that timer started
	select {
	case t.notifyChannel <- true:
	default:
	}

	return slot, nil
}

// Stop stops the current timer
func (t *Timer) Stop(db *Database) (*models.TimeSlot, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.activeSlot == nil || !t.activeSlot.IsActive() {
		return nil, nil
	}

	now := time.Now()
	err := db.StopTimeSlot(t.activeSlot.ID, now)
	if err != nil {
		return nil, err
	}

	stoppedSlot := t.activeSlot
	t.activeSlot = nil
	t.isRunning = false

	// Notify that timer stopped
	select {
	case t.notifyChannel <- false:
	default:
	}

	return stoppedSlot, nil
}

// GetActiveSlot returns the currently active time slot
func (t *Timer) GetActiveSlot() *models.TimeSlot {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.activeSlot
}

// IsRunning returns whether the timer is currently running
func (t *Timer) IsRunning() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.isRunning
}

// GetElapsedTime returns the elapsed time since the timer started
func (t *Timer) GetElapsedTime() time.Duration {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if !t.isRunning || t.activeSlot == nil {
		return 0
	}
	return time.Since(t.startTime)
}

// LoadActiveSlot loads the active slot from database
func (t *Timer) LoadActiveSlot(db *Database) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	slot, err := db.GetActiveTimeSlot()
	if err != nil {
		return err
	}

	if slot != nil {
		t.activeSlot = slot
		t.isRunning = true
		t.startTime = slot.StartTime
	} else {
		t.activeSlot = nil
		t.isRunning = false
	}

	return nil
}

