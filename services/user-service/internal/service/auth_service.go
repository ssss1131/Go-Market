package service

import (
	"GoUser/internal/domain"
	"GoUser/internal/repo"
	"GoUser/pkg/hash"
	jwtutil "GoUser/pkg/jwt"
	"errors"
	"time"

	"gorm.io/gorm"
)

type RegisterInput struct {
	Name, Surname, Email, Password string
}

type LoginInput struct {
	Email, Password string
}

type RegisterOutput struct {
	UserID uint
	Status domain.UserStatus
}

type LoginOutput struct {
	AccessToken string
	UserID      uint
	Email       string
}

type AuthService struct {
	users  *repo.Users
	signer *jwtutil.Signer
}

var (
	errEmailTaken         = errors.New("email_taken")
	errInvalidCredentials = errors.New("invalid_credentials")
)

func IsEmailTaken(err error) bool         { return errors.Is(err, errEmailTaken) }
func IsInvalidCredentials(err error) bool { return errors.Is(err, errInvalidCredentials) }

func NewAuthService(users *repo.Users, signer *jwtutil.Signer) *AuthService {
	return &AuthService{users: users, signer: signer}
}

func (s *AuthService) Register(in RegisterInput) (RegisterOutput, error) {
	pwHash, err := hash.HashPassword(in.Password)
	if err != nil {
		return RegisterOutput{}, err
	}
	user := domain.User{
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
	return RegisterOutput{user.ID, user.Status}, nil
}

func (s *AuthService) Login(in LoginInput) (LoginOutput, error) {
	u, err := s.users.ByEmail(in.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return LoginOutput{}, errInvalidCredentials
		}
		return LoginOutput{}, err
	}

	if err := hash.CheckPassword(u.PasswordHash, in.Password); err != nil {
		return LoginOutput{}, errInvalidCredentials
	}

	token, _, err := s.signer.NewAccess(u.ID, u.Email)
	if err != nil {
		return LoginOutput{}, err
	}

	return LoginOutput{
		AccessToken: token,
		UserID:      u.ID,
		Email:       u.Email,
	}, nil
}
