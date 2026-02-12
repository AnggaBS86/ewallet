package service

import (
	"errors"
	"time"

	"ewallet/internal/dto"
	"ewallet/internal/models"
	"ewallet/internal/repository"
	"ewallet/internal/utils"
)

type authService struct {
	users        repository.UserRepository
	revokedToken repository.RevokedTokenRepository
	jwtSecret    string
}

func NewAuthService(users repository.UserRepository, revokedToken repository.RevokedTokenRepository, jwtSecret string) AuthService {
	return &authService{users: users, revokedToken: revokedToken, jwtSecret: jwtSecret}
}

func (s *authService) Register(req dto.RegisterRequest) (*dto.UserResponse, error) {
	if _, err := s.users.FindByEmail(req.Email); err == nil {
		return nil, ErrConflict
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hash,
	}
	if err := s.users.CreateWithWallet(user); err != nil {
		return nil, err
	}

	return &dto.UserResponse{ID: user.ID, Name: user.Name, Email: user.Email}, nil
}

func (s *authService) Login(req dto.LoginRequest) (*dto.AuthTokenResponse, error) {
	user, err := s.users.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUnauthorized
		}
		return nil, err
	}

	if err := utils.CheckPassword(user.PasswordHash, req.Password); err != nil {
		return nil, ErrUnauthorized
	}

	token, err := utils.GenerateToken(user.ID, s.jwtSecret, 24*time.Hour)
	if err != nil {
		return nil, err
	}

	return &dto.AuthTokenResponse{Token: token}, nil
}

func (s *authService) Logout(token string) error {
	return s.revokedToken.Revoke(token)
}

func (s *authService) IsTokenRevoked(token string) (bool, error) {
	return s.revokedToken.IsRevoked(token)
}
