package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
)

// NotificationLogRepository defines the interface for notification logging
type NotificationLogRepository interface {
	Create(ctx context.Context, log *NotificationLog) error
	UpdateStatus(ctx context.Context, id int64, status string, errorMessage *string, sentAt *time.Time) error
	GetRecent(ctx context.Context, limit int) ([]*NotificationLog, error)
	GetByType(ctx context.Context, notificationType string, limit int) ([]*NotificationLog, error)
}

// NotificationLog represents a notification log entry
type NotificationLog struct {
	ID               int64           `db:"id"`
	NotificationType string          `db:"notification_type"`
	Status           string          `db:"status"`
	Message          sql.NullString  `db:"message"`
	Metadata         json.RawMessage `db:"metadata"`
	ErrorMessage     sql.NullString  `db:"error_message"`
	Attempts         int             `db:"attempts"`
	SentAt           *time.Time      `db:"sent_at"`
	CreatedAt        time.Time       `db:"created_at"`
	UpdatedAt        time.Time       `db:"updated_at"`
}

// notificationLogRepository implements NotificationLogRepository
type notificationLogRepository struct {
	db boil.ContextExecutor
}

// NewNotificationLogRepository creates a new notification log repository
func NewNotificationLogRepository(db boil.ContextExecutor) NotificationLogRepository {
	return &notificationLogRepository{
		db: db,
	}
}

// Create creates a new notification log entry
func (r *notificationLogRepository) Create(ctx context.Context, log *NotificationLog) error {
	query := `
		INSERT INTO notification_logs (
			notification_type, status, message, metadata, 
			error_message, attempts, sent_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		) RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		log.NotificationType,
		log.Status,
		log.Message,
		log.Metadata,
		log.ErrorMessage,
		log.Attempts,
		log.SentAt,
	).Scan(&log.ID)

	if err != nil {
		return err
	}

	return nil
}

// UpdateStatus updates the status of a notification log
func (r *notificationLogRepository) UpdateStatus(ctx context.Context, id int64, status string, errorMessage *string, sentAt *time.Time) error {
	query := `
		UPDATE notification_logs 
		SET status = $2, 
		    error_message = $3,
		    sent_at = $4,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id, status, errorMessage, sentAt)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("notification log not found: %d", id)
	}
	return err
}

// GetRecent retrieves recent notification logs
func (r *notificationLogRepository) GetRecent(ctx context.Context, limit int) ([]*NotificationLog, error) {
	query := `
		SELECT id, notification_type, status, message, metadata,
		       error_message, attempts, sent_at, created_at, updated_at
		FROM notification_logs
		ORDER BY created_at DESC
		LIMIT $1`

	logs := []*NotificationLog{}
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		log := &NotificationLog{}
		err := rows.Scan(
			&log.ID,
			&log.NotificationType,
			&log.Status,
			&log.Message,
			&log.Metadata,
			&log.ErrorMessage,
			&log.Attempts,
			&log.SentAt,
			&log.CreatedAt,
			&log.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return logs, nil
}

// GetByType retrieves notification logs by type
func (r *notificationLogRepository) GetByType(ctx context.Context, notificationType string, limit int) ([]*NotificationLog, error) {
	query := `
		SELECT id, notification_type, status, message, metadata,
		       error_message, attempts, sent_at, created_at, updated_at
		FROM notification_logs
		WHERE notification_type = $1
		ORDER BY created_at DESC
		LIMIT $2`

	logs := []*NotificationLog{}
	rows, err := r.db.QueryContext(ctx, query, notificationType, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		log := &NotificationLog{}
		err := rows.Scan(
			&log.ID,
			&log.NotificationType,
			&log.Status,
			&log.Message,
			&log.Metadata,
			&log.ErrorMessage,
			&log.Attempts,
			&log.SentAt,
			&log.CreatedAt,
			&log.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return logs, nil
}