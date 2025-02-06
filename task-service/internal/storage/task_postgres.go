package storage

import (
	"context"
	"fmt"
	ts "github.com/Task-bot/task-service/internal/service"
	"github.com/jackc/pgx/v5"
	"time"
)

type TaskPostgresStorage struct {
	DB *pgx.Conn
}

func NewPostgresStorage(connectionStr string) (*TaskPostgresStorage, error) {
	connection, err := pgx.Connect(context.Background(), connectionStr)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to db: %v", err)
	}

	return &TaskPostgresStorage{connection}, nil
}

func (s *TaskPostgresStorage) CreateTask(ctx context.Context, userID int64, text string, priority int32, deadline time.Time) (*ts.Task, error) {
	query := `
		INSERT INTO tasks (user_id, task_text, priority, deadline, progress)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, task_text, priority, deadline, progress;
	`
	var task ts.Task
	err := s.DB.QueryRow(ctx, query, userID, text, priority, deadline, 0).Scan(&task.ID, &task.UserID, &task.Text, &task.Priority, &task.Deadline, &task.Progress)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %v", err)
	}
	return &task, nil
}

func (s *TaskPostgresStorage) GetTasks(ctx context.Context, userID int64) ([]ts.Task, error) {
	query := `SELECT id, text, priority, deadline, progress FROM tasks WHERE user_id = $1`
	rows, err := s.DB.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %v", err)
	}
	defer rows.Close()

	var tasks []ts.Task
	for rows.Next() {
		var t ts.Task
		err := rows.Scan(&t.ID, &t.Text, &t.Priority, &t.Deadline, &t.Progress)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (s *TaskPostgresStorage) DeleteTask(ctx context.Context, taskID int64) error {
	query := `DELETE FROM tasks WHERE id = $1`
	_, err := s.DB.Exec(ctx, query, taskID)
	if err != nil {
		return fmt.Errorf("failed to delete task: %v", err)
	}
	return nil
}

func (s *TaskPostgresStorage) UpdateTaskStatus(ctx context.Context, taskID int64, progress int32) error {
	query := `UPDATE tasks SET progress = $1 WHERE id = $2`
	_, err := s.DB.Exec(ctx, query, progress, taskID)
	if err != nil {
		return fmt.Errorf("failed to update task progress: %v", err)
	}
	return nil
}
