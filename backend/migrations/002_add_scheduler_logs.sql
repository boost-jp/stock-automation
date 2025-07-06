-- Create scheduler_logs table
CREATE TABLE IF NOT EXISTS scheduler_logs (
    id SERIAL PRIMARY KEY,
    task_name VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL,
    started_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    error_message TEXT,
    duration_ms INT,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX idx_scheduler_logs_task_name ON scheduler_logs(task_name);
CREATE INDEX idx_scheduler_logs_status ON scheduler_logs(status);
CREATE INDEX idx_scheduler_logs_started_at ON scheduler_logs(started_at);