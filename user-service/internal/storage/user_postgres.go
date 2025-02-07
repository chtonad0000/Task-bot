package storage

import (
	"context"
	"fmt"
	us "github.com/Task-bot/user-service/internal/service"
	"github.com/jackc/pgx/v5"
)

type UserPostgresStorage struct {
	DB *pgx.Conn
}

func NewPostgresStorage(connectionStr string) (*UserPostgresStorage, error) {
	connection, err := pgx.Connect(context.Background(), connectionStr)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to db: %v", err)
	}

	return &UserPostgresStorage{connection}, nil
}

func (s *UserPostgresStorage) CreateUser(ctx context.Context, username string, tgUserId int64) (*us.User, error) {
	query := `
		INSERT INTO users (username, tg_user_id)
		VALUES ($1, $2)
		RETURNING id, username, tg_user_id, created_at;
	`
	var user us.User
	err := s.DB.QueryRow(ctx, query, username, tgUserId).Scan(&user.ID, &user.Username, &user.TgUserId, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}
	return &user, nil
}

func (s *UserPostgresStorage) GetUser(ctx context.Context, tgUserId int64) (*us.User, error) {
	query := `SELECT id, username, tg_user_id, created_at FROM users WHERE tg_user_id = $1`
	var user us.User
	err := s.DB.QueryRow(ctx, query, tgUserId).Scan(&user.ID, &user.Username, &user.TgUserId, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}
	return &user, nil
}
