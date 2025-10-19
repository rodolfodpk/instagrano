package service

import (
    "errors"
    "github.com/rodolfodpk/instagrano/internal/domain"
    "github.com/rodolfodpk/instagrano/internal/repository/postgres"
    "golang.org/x/crypto/bcrypt"
)

var ErrInvalidInput = errors.New("invalid input")

type AuthService struct {
    userRepo postgres.UserRepository
}

func NewAuthService(userRepo postgres.UserRepository) *AuthService {
    return &AuthService{userRepo: userRepo}
}

func (s *AuthService) Register(username, email, password string) (*domain.User, error) {
    if username == "" || email == "" || password == "" {
        return nil, ErrInvalidInput
    }

    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return nil, err
    }

    user := &domain.User{
        Username: username,
        Email:    email,
        Password: string(hashedPassword),
    }

    if err := s.userRepo.Create(user); err != nil {
        return nil, err
    }

    return user, nil
}
