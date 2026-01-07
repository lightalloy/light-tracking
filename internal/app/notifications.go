package app

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"time"
)

type NotificationManager struct {
	app            *App
	ctx            context.Context
	lastNotifyTime time.Time
	notifyInterval time.Duration // Notify every 2 hours
}

// NewNotificationManager creates a new notification manager
func NewNotificationManager(app *App) *NotificationManager {
	return &NotificationManager{
		app:            app,
		notifyInterval: 2 * time.Hour,
		lastNotifyTime: time.Time{},
	}
}

// Start starts monitoring for long sessions and sends notifications
func (n *NotificationManager) Start(ctx context.Context) {
	n.ctx = ctx
	go n.monitorLongSessions()
}

// monitorLongSessions checks if timer is running for a long time and sends notifications
func (n *NotificationManager) monitorLongSessions() {
	ticker := time.NewTicker(5 * time.Minute) // Check every 5 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if n.app.IsTimerRunning() {
				elapsed := n.app.GetElapsedTime()
				elapsedDuration := time.Duration(elapsed) * time.Second

				// Send notification if session is longer than notifyInterval
				// and we haven't notified recently
				if elapsedDuration >= n.notifyInterval {
					timeSinceLastNotify := time.Since(n.lastNotifyTime)
					if timeSinceLastNotify >= n.notifyInterval {
						activeSlot := n.app.GetActiveTimeSlot()
						if activeSlot != nil {
							n.SendNotification(
								"Long Session Alert",
								"You've been working on '"+activeSlot.TaskName+"' for "+formatDuration(elapsedDuration),
							)
							n.lastNotifyTime = time.Now()
						}
					}
				}
			}
		case <-n.ctx.Done():
			return
		}
	}
}

// SendNotification sends a desktop notification
func (n *NotificationManager) SendNotification(title, message string) error {
	switch runtime.GOOS {
	case "linux":
		return n.sendLinuxNotification(title, message)
	case "darwin":
		return n.sendMacOSNotification(title, message)
	case "windows":
		return n.sendWindowsNotification(title, message)
	default:
		return nil
	}
}

// sendLinuxNotification sends a notification on Linux using notify-send or dbus
func (n *NotificationManager) sendLinuxNotification(title, message string) error {
	// Try notify-send first (most common)
	cmd := exec.Command("notify-send", title, message, "--app-name=Light Tracking")
	if err := cmd.Run(); err == nil {
		return nil
	}

	// Fallback to dbus-send
	cmd = exec.Command("dbus-send", "--type=method_call",
		"--dest=org.freedesktop.Notifications",
		"/org/freedesktop/Notifications",
		"org.freedesktop.Notifications.Notify",
		"string:Light Tracking",
		"uint32:0",
		"string:",
		"string:"+title,
		"string:"+message,
		"array:string:",
		"dict:string:",
		"int32:5000")
	return cmd.Run()
}

// sendMacOSNotification sends a notification on macOS
func (n *NotificationManager) sendMacOSNotification(title, message string) error {
	script := `display notification "` + message + `" with title "` + title + `"`
	cmd := exec.Command("osascript", "-e", script)
	return cmd.Run()
}

// sendWindowsNotification sends a notification on Windows
func (n *NotificationManager) sendWindowsNotification(title, message string) error {
	// Windows 10+ has built-in toast notifications
	// This is a simplified version - in production you might want to use
	// a proper Windows notification library
	cmd := exec.Command("powershell", "-Command",
		`[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null; `+
			`[Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime] | Out-Null; `+
			`$xml = [Windows.Data.Xml.Dom.XmlDocument]::new(); `+
			`$xml.LoadXml('<toast><visual><binding template="ToastText02"><text id="1">`+title+`</text><text id="2">`+message+`</text></binding></visual></toast>'); `+
			`$toast = [Windows.UI.Notifications.ToastNotification]::new($xml); `+
			`[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier("Light Tracking").Show($toast)`)
	return cmd.Run()
}

// formatDuration formats a duration as "X hours Y minutes"
func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	if hours > 0 {
		if minutes > 0 {
			return formatPlural(hours, "hour") + " and " + formatPlural(minutes, "minute")
		}
		return formatPlural(hours, "hour")
	}
	return formatPlural(minutes, "minute")
}

func formatPlural(n int, unit string) string {
	if n == 1 {
		return "1 " + unit
	}
	return fmt.Sprintf("%d %ss", n, unit)
}
