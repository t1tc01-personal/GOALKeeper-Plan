package service

import (
	"goalkeeper-plan/internal/user/model"
	"goalkeeper-plan/internal/user/repository"
)

type UserConfiguration func(s *userService) error

type UserService interface {
	CreateUser(email, name string) (*model.User, error)
	GetUserByID(id string) (*model.User, error)
	GetUserByEmail(email string) (*model.User, error)
	UpdateUser(user *model.User) error
	DeleteUser(id string) error
	ListUsers(limit, offset int) ([]*model.User, error)
}

type userService struct {
	userRepository repository.UserRepository
}

func NewUserService(configs ...UserConfiguration) (UserService, error) {
	s := &userService{}
	for _, cfg := range configs {
		err := cfg(s)
		if err != nil {
			return nil, err
		}
	}
	return s, nil
}

func WithUserRepository(r repository.UserRepository) UserConfiguration {
	return func(s *userService) error {
		s.userRepository = r
		return nil
	}
}

func (s *userService) CreateUser(email, name string) (*model.User, error) {
	user := &model.User{
		Email: email,
		Name:  name,
	}
	err := s.userRepository.Create(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) GetUserByID(id string) (*model.User, error) {
	return s.userRepository.GetByID(id)
}

func (s *userService) GetUserByEmail(email string) (*model.User, error) {
	return s.userRepository.GetByEmail(email)
}

func (s *userService) UpdateUser(user *model.User) error {
	return s.userRepository.Update(user)
}

func (s *userService) DeleteUser(id string) error {
	return s.userRepository.Delete(id)
}

func (s *userService) ListUsers(limit, offset int) ([]*model.User, error) {
	return s.userRepository.List(limit, offset)
}

