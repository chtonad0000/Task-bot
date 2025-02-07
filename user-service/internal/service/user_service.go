package service

import (
	"context"
	pb "github.com/Task-bot/user-service/internal/generated/user"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type User struct {
	ID        int64
	Username  string
	TgUserId  []byte
	CreatedAt time.Time
}

type UserStorage interface {
	CreateUser(ctx context.Context, username string, tgUserId []byte) (*User, error)
	GetUser(ctx context.Context, tgUserId []byte) (*User, error)
}

type UserService struct {
	storage UserStorage
	pb.UnimplementedUserServiceServer
}

func NewUserService(storage UserStorage) *UserService {
	return &UserService{storage: storage}
}

func (u *UserService) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
	user, err := u.storage.CreateUser(ctx, req.Username, req.TgUserId)
	if err != nil {
		return nil, err
	}
	return &pb.UserResponse{
		Id:        user.ID,
		Username:  user.Username,
		TgUserId:  user.TgUserId,
		CreatedAt: timestamppb.New(user.CreatedAt),
	}, nil
}

func (u *UserService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UserResponse, error) {
	user, err := u.storage.GetUser(ctx, req.TgUserId)
	if err != nil {
		return nil, err
	}
	return &pb.UserResponse{
		Id:        user.ID,
		Username:  user.Username,
		TgUserId:  user.TgUserId,
		CreatedAt: timestamppb.New(user.CreatedAt),
	}, nil
}
