package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
)

// SchedulerLogRepository defines the interface for scheduler logging
type SchedulerLogRepository interface {
	StartTask(ctx context.Context, taskName string) (int64, error)
	CompleteTask(ctx context.Context, id int64, duration time.Duration) error
	FailTask(ctx context.Context, id int64, duration time.Duration, err error) error
	GetRecentLogs(ctx context.Context, limit int) ([]*SchedulerLog, error)
	GetTaskLogs(ctx context.Context, taskName string, limit int) ([]*SchedulerLog, error)
}

// SchedulerLog represents a scheduler log entry
type SchedulerLog struct {
	ID           int64           `db:"id"`
	TaskName     string          `db:"task_name"`
	Status       string          `db:"status"`
	StartedAt    time.Time       `db:"started_at"`
	CompletedAt  *time.Time      `db:"completed_at"`
	ErrorMessage sql.NullString  `db:"error_message"`
	DurationMs   sql.NullInt64   `db:"duration_ms"`
	Metadata     json.RawMessage `db:"metadata"`
	CreatedAt    time.Time       `db:"created_at"`
}

// schedulerLogRepository implements SchedulerLogRepository
type schedulerLogRepository struct {
	db boil.ContextExecutor
}

// NewSchedulerLogRepository creates a new scheduler log repository
func NewSchedulerLogRepository(db boil.ContextExecutor) SchedulerLogRepository {
	return &schedulerLogRepository{
		db: db,
	}
}

// StartTask logs the start of a scheduled task
func (r *schedulerLogRepository) StartTask(ctx context.Context, taskName string) (int64, error) {
	query := `
		INSERT INTO scheduler_logs (task_name, status, started_at)
		VALUES ($1, $2, $3)
		RETURNING id`

	var id int64
	err := r.db.QueryRowContext(ctx, query, taskName, "running", time.Now()).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// CompleteTask marks a task as completed
func (r *schedulerLogRepository) CompleteTask(ctx context.Context, id int64, duration time.Duration) error {
	query := `
		UPDATE scheduler_logs 
		SET status = 'completed',
		    completed_at = $2,
		    duration_ms = $3
		WHERE id = $1`

	completedAt := time.Now()
	durationMs := duration.Milliseconds()

	_, err := r.db.ExecContext(ctx, query, id, completedAt, durationMs)
	return err
}

// FailTask marks a task as failed
func (r *schedulerLogRepository) FailTask(ctx context.Context, id int64, duration time.Duration, err error) error {
	query := `
		UPDATE scheduler_logs 
		SET status = 'failed',
		    completed_at = $2,
		    duration_ms = $3,
		    error_message = $4
		WHERE id = $1`

	completedAt := time.Now()
	durationMs := duration.Milliseconds()
	errMsg := err.Error()

	_, execErr := r.db.ExecContext(ctx, query, id, completedAt, durationMs, errMsg)
	return execErr
}

// GetRecentLogs retrieves recent scheduler logs
func (r *schedulerLogRepository) GetRecentLogs(ctx context.Context, limit int) ([]*SchedulerLog, error) {
	query := `
		SELECT id, task_name, status, started_at, completed_at,
		       error_message, duration_ms, metadata, created_at
		FROM scheduler_logs
		ORDER BY started_at DESC
		LIMIT $1`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := []*SchedulerLog{}
	for rows.Next() {
		log := &SchedulerLog{}
		err := rows.Scan(
			&log.ID,
			&log.TaskName,
			&log.Status,
			&log.StartedAt,
			&log.CompletedAt,
			&log.ErrorMessage,
			&log.DurationMs,
			&log.Metadata,
			&log.CreatedAt,
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

// GetTaskLogs retrieves logs for a specific task
func (r *schedulerLogRepository) GetTaskLogs(ctx context.Context, taskName string, limit int) ([]*SchedulerLog, error) {
	query := `
		SELECT id, task_name, status, started_at, completed_at,
		       error_message, duration_ms, metadata, created_at
		FROM scheduler_logs
		WHERE task_name = $1
		ORDER BY started_at DESC
		LIMIT $2`

	rows, err := r.db.QueryContext(ctx, query, taskName, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := []*SchedulerLog{}
	for rows.Next() {
		log := &SchedulerLog{}
		err := rows.Scan(
			&log.ID,
			&log.TaskName,
			&log.Status,
			&log.StartedAt,
			&log.CompletedAt,
			&log.ErrorMessage,
			&log.DurationMs,
			&log.Metadata,
			&log.CreatedAt,
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

