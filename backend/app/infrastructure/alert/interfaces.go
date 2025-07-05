package alert

import "context"

// Service defines the interface for alert notifications
type Service interface {
	// Send sends an alert
	Send(ctx context.Context, alert *Alert) error

	// SendCritical sends a critical alert
	SendCritical(ctx context.Context, title, message string, err error) error

	// SendError sends an error alert
	SendError(ctx context.Context, title, message string, err error) error

	// SendWarning sends a warning alert
	SendWarning(ctx context.Context, title, message string, err error) error
}

