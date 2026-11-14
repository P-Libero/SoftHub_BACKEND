package service

import (
	"context"
	"errors"
	"softhub/model"
	"softhub/pkg/utils"
	"softhub/repository"
	"softhub/request"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	Repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		Repo: repo,
	}
}

// Register 注册：查重 -> 加密 -> 落库

func (s *UserService) Register(ctx context.Context, req request.RegisterReq) error {
	if err := s.Repo.CheckUsernameExist(ctx, req.Username); err != nil {
		if errors.Is(err, repository.ErrUsernameExists) {
			return ErrUsernameTaken
		}
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	req.Password = string(hash)
	user := &model.User{
		Username:     req.Username,
		PasswordHash: req.Password,
		Role:         "user",
	}
	if err := s.Repo.CreateUser(ctx, user); err != nil {
		if errors.Is(err, repository.ErrUsernameExists) {
			return ErrUsernameTaken
		}
		return err
	}
	return nil
}

// Login 登录：查用户 -> 比对密码 -> 签发 Token

func (s *UserService) Login(ctx context.Context, req request.LoginReq) (string, error) {
	user, err := s.Repo.FindUserByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return "", ErrInvalidCredentials
	}
	return utils.GenerateToken(user.ID, user.Username, user.Role)
}

// GetProfile 根据用户 ID 查询用户信息

func (s *UserService) GetProfile(ctx context.Context, id uint64) (*model.User, error) {
	user, err := s.Repo.FindUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (s *UserService) EnsureAdmin(ctx context.Context, username, password string) error {
	err := s.Repo.CheckUsernameExist(ctx, username)
	if err != nil {
		if errors.Is(err, repository.ErrUsernameExists) {
			return nil
		}
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.Repo.CreateUser(ctx, &model.User{
		Username:     username,
		PasswordHash: string(hash),
		Role:         "admin",
	})
}
