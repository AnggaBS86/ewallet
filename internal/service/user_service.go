package service

import (
	"errors"

	"ewallet/internal/dto"
	"ewallet/internal/repository"
)

type userService struct {
	users repository.UserRepository
}

func NewUserService(users repository.UserRepository) UserService {
	return &userService{users: users}
}

func (s *userService) Profile(userID uint) (*dto.UserResponse, error) {
	user, err := s.users.FindByID(userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &dto.UserResponse{ID: user.ID, Name: user.Name, Email: user.Email}, nil
}
