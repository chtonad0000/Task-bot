package storage

import (
	"context"
	"fmt"
	"github.com/Task-bot/notification-service/internal/service"
	"time"

	"github.com/jackc/pgx/v5"
)

type NotificationPostgresStorage struct {
	DB *pgx.Conn
}

func NewPostgresStorage(connectionStr string) (*NotificationPostgresStorage, error) {
	connection, err := pgx.Connect(context.Background(), connectionStr)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to db: %v", err)
	}
	return &NotificationPostgresStorage{connection}, nil
}

func (s *NotificationPostgresStorage) CreateNotification(ctx context.Context, userID int64, message string, notifyTime time.Time) (*service.Notification, error) {
	query := `
		INSERT INTO notifications (user_id, message, notify_time)
		
		VALUES ($1, $2, $3)
		RETURNING notification_id, user_id, message, notify_time;
	`
	var notification service.Notification
	err := s.DB.QueryRow(ctx, query, userID, message, notifyTime).Scan(&notification.ID, &notification.UserID, &notification.Message, &notification.NotifyTime)
	if err != nil {
		return nil, fmt.Errorf("failed to create notification: %v", err)
	}
	return &notification, nil
}

func (s *NotificationPostgresStorage) GetDueNotifications(ctx context.Context, now time.Time) ([]service.Notification, error) {
	query := `SELECT notification_id, user_id, message, notify_time FROM notifications WHERE notify_time <= $1`
	rows, err := s.DB.Query(ctx, query, now)
	if err != nil {
		return nil, fmt.Errorf("failed to get due notifications: %v", err)
	}
	defer rows.Close()

	var notifications []service.Notification
	for rows.Next() {
		var n service.Notification
		err := rows.Scan(&n.ID, &n.UserID, &n.Message, &n.NotifyTime)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		notifications = append(notifications, n)
	}
	return notifications, nil
}

func (s *NotificationPostgresStorage) DeleteNotification(ctx context.Context, notificationID int64) error {
	query := `DELETE FROM notifications WHERE notification_id = $1`
	_, err := s.DB.Exec(ctx, query, notificationID)
	if err != nil {
		return fmt.Errorf("failed to delete notification: %v", err)
	}
	return nil
}
