package gui

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// ActivityLogger handles thread-safe log output to GUI widgets and terminal.
type ActivityLogger struct {
	mu         sync.Mutex
	lines      []string
	maxLines   int
	logWidget  *widget.Entry
	statusText *widget.Label
}

func NewActivityLogger(statusLabel *widget.Label) *ActivityLogger {
	entry := widget.NewMultiLineEntry()
	entry.Wrapping = fyne.TextWrapBreak
	entry.TextStyle = fyne.TextStyle{Monospace: true}

	return &ActivityLogger{
		lines:      make([]string, 0, 100),
		maxLines:   500,
		logWidget:  entry,
		statusText: statusLabel,
	}
}

func (l *ActivityLogger) Widget() *widget.Entry {
	return l.logWidget
}

func (l *ActivityLogger) Log(msg string) {
	timestamp := time.Now().Format("15:04:05")
	formatted := fmt.Sprintf("[%s] %s", timestamp, msg)

	l.mu.Lock()
	l.lines = append(l.lines, formatted)
	if len(l.lines) > l.maxLines {
		l.lines = l.lines[len(l.lines)-l.maxLines:]
	}
	allText := strings.Join(l.lines, "\n")
	l.mu.Unlock()

	// Update GUI on main thread / safe queue
	fyne.Do(func() {
		l.logWidget.SetText(allText)
		l.logWidget.CursorRow = len(l.lines)
		if l.statusText != nil {
			l.statusText.SetText(msg)
		}
	})

	fmt.Println(formatted)
}

func (l *ActivityLogger) Clear() {
	l.mu.Lock()
	l.lines = l.lines[:0]
	l.mu.Unlock()

	fyne.Do(func() {
		l.logWidget.SetText("")
		if l.statusText != nil {
			l.statusText.SetText("Logs cleared")
		}
	})
}
