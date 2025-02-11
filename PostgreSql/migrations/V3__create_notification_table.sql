CREATE TABLE IF NOT EXISTS notifications (
    notification_id SERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    message TEXT NOT NULL,
    notify_time TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_tasks_id ON notifications(user_id);