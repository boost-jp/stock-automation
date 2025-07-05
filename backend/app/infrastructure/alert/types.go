package alert

import "time"

// Level represents the severity level of an alert
type Level int

const (
	// LevelCritical represents critical errors that need immediate attention
	LevelCritical Level = iota
	// LevelError represents errors that should be addressed
	LevelError
	// LevelWarning represents warnings that might need attention
	LevelWarning
)

// String returns the string representation of the alert level
func (l Level) String() string {
	switch l {
	case LevelCritical:
		return "CRITICAL"
	case LevelError:
		return "ERROR"
	case LevelWarning:
		return "WARNING"
	default:
		return "UNKNOWN"
	}
}

// Alert represents an error alert
type Alert struct {
	Level     Level
	Title     string
	Message   string
	Error     error
	Timestamp time.Time
	Context   map[string]interface{}
}

// NewAlert creates a new alert
func NewAlert(level Level, title, message string, err error) *Alert {
	return &Alert{
		Level:     level,
		Title:     title,
		Message:   message,
		Error:     err,
		Timestamp: time.Now(),
		Context:   make(map[string]interface{}),
	}
}

// WithContext adds context to the alert
func (a *Alert) WithContext(key string, value interface{}) *Alert {
	a.Context[key] = value
	return a
}


