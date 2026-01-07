package app

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/getlantern/systray"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type SystrayManager struct {
	app          *App
	ctx          context.Context
	mu           sync.RWMutex
	isRunning    bool
	showItem     *systray.MenuItem
	hideItem     *systray.MenuItem
	quitItem     *systray.MenuItem
	statusItem   *systray.MenuItem
	iconActive   []byte
	iconInactive []byte
}

// NewSystrayManager creates a new systray manager
func NewSystrayManager(app *App) *SystrayManager {
	return &SystrayManager{
		app: app,
	}
}

// Run starts the systray in a separate goroutine
func (s *SystrayManager) Run(ctx context.Context) {
	s.ctx = ctx
	// Load icons before starting systray
	s.loadIcons()
	// Start systray in a goroutine (required for Wails)
	go systray.Run(s.onReady, s.onExit)
}

// loadIcons loads icons from files or creates default ones
func (s *SystrayManager) loadIcons() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Try to load separate icons for active/inactive states
	// First, try build/icons directory (preferred)
	activePath := "build/icons/icon-active.png"
	inactivePath := "build/icons/icon-inactive.png"

	// If not found, try relative to executable
	if _, err := os.Stat(activePath); os.IsNotExist(err) {
		exe, err := os.Executable()
		if err == nil {
			exeDir := filepath.Dir(exe)
			activePath = filepath.Join(exeDir, "build", "icons", "icon-active.png")
			inactivePath = filepath.Join(exeDir, "build", "icons", "icon-inactive.png")
			if _, err := os.Stat(activePath); os.IsNotExist(err) {
				// Try parent directory
				activePath = filepath.Join(filepath.Dir(exeDir), "build", "icons", "icon-active.png")
				inactivePath = filepath.Join(filepath.Dir(exeDir), "build", "icons", "icon-inactive.png")
			}
		}
	}

	// Try to load active icon
	activeBytes, err := os.ReadFile(activePath)
	if err != nil {
		activeBytes = nil
	}

	// Try to load inactive icon
	inactiveBytes, err := os.ReadFile(inactivePath)
	if err != nil {
		inactiveBytes = nil
	}

	// If both icons found, use them
	if activeBytes != nil && inactiveBytes != nil {
		s.iconActive = activeBytes
		s.iconInactive = inactiveBytes
		return
	}

	// Fallback: try to use single appicon.png and create variants
	iconPath := "build/appicon.png"
	if _, err := os.Stat(iconPath); os.IsNotExist(err) {
		exe, err := os.Executable()
		if err == nil {
			exeDir := filepath.Dir(exe)
			iconPath = filepath.Join(exeDir, "build", "appicon.png")
			if _, err := os.Stat(iconPath); os.IsNotExist(err) {
				iconPath = filepath.Join(filepath.Dir(exeDir), "build", "appicon.png")
			}
		}
	}

	iconBytes, err := os.ReadFile(iconPath)
	if err != nil {
		// Use default icons if file not found
		s.iconActive = s.createDefaultIcon(true)
		s.iconInactive = s.createDefaultIcon(false)
	} else {
		// Use same icon for both states (will be updated when separate icons are added)
		s.iconActive = iconBytes
		s.iconInactive = iconBytes
	}
}

// createDefaultIcon creates a visual PNG icon with a circle
func (s *SystrayManager) createDefaultIcon(active bool) []byte {
	const size = 32
	const center = size / 2
	const radius = 12.0

	// Create RGBA image with transparent background
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// Define colors
	var circleColor color.RGBA
	if active {
		// Green color for active timer: RGB(76, 175, 80)
		circleColor = color.RGBA{76, 175, 80, 255}
	} else {
		// Dark gray for inactive timer outline
		circleColor = color.RGBA{100, 100, 100, 255}
	}

	// Draw circle
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - center
			dy := float64(y) - center
			distance := math.Sqrt(dx*dx + dy*dy)

			if active {
				// Filled circle for active state
				if distance <= radius {
					img.Set(x, y, circleColor)
				}
			} else {
				// Outline circle for inactive state
				// Draw pixels that are on the circle outline (with some thickness)
				if distance >= radius-1.5 && distance <= radius+0.5 {
					img.Set(x, y, circleColor)
				}
			}
		}
	}

	// Encode to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		// Fallback to minimal PNG if encoding fails
		return []byte{
			0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D,
			0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
			0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4, 0x89, 0x00, 0x00, 0x00,
			0x0A, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00,
			0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00, 0x00, 0x00, 0x00, 0x49,
			0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
		}
	}

	return buf.Bytes()
}

// onReady is called when systray is ready
func (s *SystrayManager) onReady() {
	s.mu.RLock()
	icon := s.iconInactive
	s.mu.RUnlock()

	// Set icon and tooltip
	if len(icon) > 0 {
		systray.SetIcon(icon)
	}
	systray.SetTooltip("Light Tracking")

	// Create menu items
	s.statusItem = systray.AddMenuItem("Timer: Stopped", "Current timer status")
	s.statusItem.Disable()

	systray.AddSeparator()

	s.showItem = systray.AddMenuItem("Show Window", "Show the main window")
	s.hideItem = systray.AddMenuItem("Hide Window", "Hide the main window")
	s.hideItem.Hide()

	systray.AddSeparator()

	s.quitItem = systray.AddMenuItem("Quit", "Quit the application")

	// Start monitoring timer status
	go s.monitorTimerStatus()

	// Handle menu clicks
	go s.handleMenuClicks()
}

// onExit is called when systray exits
func (s *SystrayManager) onExit() {
	// Cleanup if needed
}

// monitorTimerStatus periodically checks timer status and updates icon
func (s *SystrayManager) monitorTimerStatus() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.updateStatus()
		case <-s.ctx.Done():
			return
		}
	}
}

// updateStatus updates the systray icon and status based on timer state
func (s *SystrayManager) updateStatus() {
	isRunning := s.app.IsTimerRunning()

	s.mu.Lock()
	wasRunning := s.isRunning
	s.isRunning = isRunning
	s.mu.Unlock()

	if wasRunning != isRunning {
		s.mu.RLock()
		var icon []byte
		if isRunning {
			icon = s.iconActive
		} else {
			icon = s.iconInactive
		}
		s.mu.RUnlock()

		if len(icon) > 0 {
			systray.SetIcon(icon)
		}

		if isRunning {
			activeSlot := s.app.GetActiveTimeSlot()
			if activeSlot != nil {
				s.statusItem.SetTitle("Timer: Running - " + activeSlot.TaskName)
			} else {
				s.statusItem.SetTitle("Timer: Running")
			}
		} else {
			s.statusItem.SetTitle("Timer: Stopped")
		}
	} else if isRunning {
		// Update elapsed time in status
		activeSlot := s.app.GetActiveTimeSlot()
		if activeSlot != nil {
			elapsed := s.app.GetElapsedTime()
			hours := elapsed / 3600
			minutes := (elapsed % 3600) / 60
			seconds := elapsed % 60
			s.statusItem.SetTitle("Timer: " + activeSlot.TaskName +
				" (" + formatTime(hours, minutes, seconds) + ")")
		}
	}
}

// handleMenuClicks handles clicks on systray menu items
func (s *SystrayManager) handleMenuClicks() {
	for {
		select {
		case <-s.showItem.ClickedCh:
			runtime.WindowShow(s.ctx)
			s.showItem.Hide()
			s.hideItem.Show()
		case <-s.hideItem.ClickedCh:
			runtime.WindowHide(s.ctx)
			s.hideItem.Hide()
			s.showItem.Show()
		case <-s.quitItem.ClickedCh:
			systray.Quit()
			runtime.Quit(s.ctx)
		case <-s.ctx.Done():
			return
		}
	}
}

// formatTime formats hours, minutes, seconds as HH:MM:SS
func formatTime(hours, minutes, seconds int64) string {
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}
