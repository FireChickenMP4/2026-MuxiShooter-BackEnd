package service

import (
	"MuXi/2026-MuxiShooter-Backend/dto"
	"MuXi/2026-MuxiShooter-Backend/models"
	"errors"
	"time"
)

var (
	ErrUserAlreadyExists = errors.New("用户已存在")
	ErrUserNotFound      = errors.New("用户不存在")
	ErrInvalidPassword   = errors.New("密码错误")
)

type UserRepository interface {
	FindByUsername(username string) (*models.User, bool, error)
	Create(user *models.User) error
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, password string) error
}

type TokenService interface {
	GenerateToken(user models.User) (token string, expirationTime time.Time, err error)
}

type AuthService struct {
	userRepository       UserRepository
	passwordHasher       PasswordHasher
	tokenService         TokenService
	defaultHeadImagePath string
}

func NewAuthService(userRepository UserRepository, passwordHasher PasswordHasher, tokenService TokenService, defaultHeadImagePath string) *AuthService {
	return &AuthService{
		userRepository:       userRepository,
		passwordHasher:       passwordHasher,
		tokenService:         tokenService,
		defaultHeadImagePath: defaultHeadImagePath,
	}
}

func (s *AuthService) Register(req dto.RegisterRequest) (dto.AuthData, error) {
	existedUser, existed, err := s.userRepository.FindByUsername(req.UserName)
	if err != nil {
		return dto.AuthData{}, err
	}
	if existed && existedUser != nil {
		return dto.AuthData{}, ErrUserAlreadyExists
	}

	hashedPsw, err := s.passwordHasher.Hash(req.Password)
	if err != nil {
		return dto.AuthData{}, err
	}

	newUser := models.User{
		Username:      req.UserName,
		Password:      hashedPsw,
		Group:         "user",
		HeadImagePath: s.defaultHeadImagePath,
	}

	if err = s.userRepository.Create(&newUser); err != nil {
		return dto.AuthData{}, err
	}

	token, expirationTime, err := s.tokenService.GenerateToken(newUser)
	if err != nil {
		return dto.AuthData{}, err
	}

	return dto.AuthData{
		User: dto.CommonUserData{
			UserID:        newUser.ID,
			Username:      newUser.Username,
			Group:         newUser.Group,
			HeadImagePath: newUser.HeadImagePath,
			StrengthCoin:  newUser.StrengthCoin,
			SelectCoin:    newUser.SelectCoin,
		},
		Token:     token,
		ExpiresAt: expirationTime.Unix(),
	}, nil
}

func (s *AuthService) Login(req dto.LoginRequest) (dto.AuthData, error) {
	user, existed, err := s.userRepository.FindByUsername(req.UserName)
	if err != nil {
		return dto.AuthData{}, err
	}
	if !existed || user == nil {
		return dto.AuthData{}, ErrUserNotFound
	}

	if err = s.passwordHasher.Compare(user.Password, req.Password); err != nil {
		return dto.AuthData{}, ErrInvalidPassword
	}

	token, expirationTime, err := s.tokenService.GenerateToken(*user)
	if err != nil {
		return dto.AuthData{}, err
	}

	return dto.AuthData{
		User: dto.CommonUserData{
			UserID:        user.ID,
			Username:      user.Username,
			Group:         user.Group,
			HeadImagePath: user.HeadImagePath,
			StrengthCoin:  user.StrengthCoin,
			SelectCoin:    user.SelectCoin,
		},
		Token:     token,
		ExpiresAt: expirationTime.Unix(),
	}, nil
}
