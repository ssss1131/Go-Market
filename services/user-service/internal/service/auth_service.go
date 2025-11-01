package service

import (
	"GoMarket/internal/domain"
	"GoMarket/internal/repo"
	"GoMarket/pkg/hash"
	"errors"
	"github.com/google/uuid"
	"time"
)

type RegisterInput struct {
	Name, Surname, Email, Password string
}

type RegisterOutput struct {
	UserID uuid.UUID
	Status domain.UserStatus
}

type AuthService struct {
	users *repo.Users
}

var errEmailTaken = errors.New("email_taken")

func IsEmailTaken(err error) bool { return errors.Is(err, errEmailTaken) }

func NewAuthService(users *repo.Users) *AuthService {
	return &AuthService{users: users}
}

func (s *AuthService) Register(in RegisterInput) (RegisterOutput, error) {
	pwHash, err := hash.HashPassword(in.Password)
	if err != nil {
		return RegisterOutput{}, err
	}
	user := domain.User{
		ID:           uuid.New(),
		Name:         in.Name,
		Surname:      in.Surname,
		Email:        in.Email,
		PasswordHash: pwHash,
		Status:       domain.StatusPending,
		CreatedAt:    time.Now(),
	}
	if err := s.users.Create(&user); err != nil {
		if repo.IsUniqueEmail(err) {
			return RegisterOutput{}, errEmailTaken
		}
		return RegisterOutput{}, err
	}
	return RegisterOutput{user.ID, user.Status}, err
}
