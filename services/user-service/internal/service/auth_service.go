package service

import (
	"GoUser/internal/domain"
	"GoUser/internal/repo"
	"GoUser/pkg/hash"
	jwtutil "GoUser/pkg/jwt"
	"GoUser/pkg/kafka"
	"GoUser/pkg/token"
	"context"
	"errors"
	"log"
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

type UserRegisteredEvent struct {
	UserID  uint   `json:"user_id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Token   string `json:"token"`
	BaseURL string `json:"base_url"`
}

type AuthService struct {
	users    *repo.Users
	signer   *jwtutil.Signer
	producer *kafka.Producer
	baseURL  string
}

var (
	errEmailTaken         = errors.New("email_taken")
	errInvalidCredentials = errors.New("invalid_credentials")
	errInvalidToken       = errors.New("invalid_token")
	errAlreadyActive      = errors.New("already_active")
)

func IsEmailTaken(err error) bool         { return errors.Is(err, errEmailTaken) }
func IsInvalidCredentials(err error) bool { return errors.Is(err, errInvalidCredentials) }
func IsInvalidToken(err error) bool       { return errors.Is(err, errInvalidToken) }
func IsAlreadyActive(err error) bool      { return errors.Is(err, errAlreadyActive) }

func NewAuthService(users *repo.Users, signer *jwtutil.Signer, producer *kafka.Producer, baseURL string) *AuthService {
	return &AuthService{
		users:    users,
		signer:   signer,
		producer: producer,
		baseURL:  baseURL,
	}
}

func (s *AuthService) Register(in RegisterInput) (RegisterOutput, error) {
	pwHash, err := hash.HashPassword(in.Password)
	if err != nil {
		return RegisterOutput{}, err
	}

	verifyToken, err := token.GenerateVerificationToken()
	if err != nil {
		return RegisterOutput{}, err
	}

	user := domain.User{
		Name:              in.Name,
		Surname:           in.Surname,
		Email:             in.Email,
		PasswordHash:      pwHash,
		Status:            domain.StatusPending,
		VerificationToken: verifyToken,
		CreatedAt:         time.Now(),
	}

	if err := s.users.Create(&user); err != nil {
		if repo.IsUniqueEmail(err) {
			return RegisterOutput{}, errEmailTaken
		}
		return RegisterOutput{}, err
	}

	event := UserRegisteredEvent{
		UserID:  user.ID,
		Email:   user.Email,
		Name:    user.Name,
		Token:   verifyToken,
		BaseURL: s.baseURL,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.producer.Send(ctx, "user.registered", user.Email, event); err != nil {
		log.Printf("Error occured while sending event to kafka %v", err)
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

	tok, _, err := s.signer.NewAccess(u.ID, u.Email, string(u.Status))
	if err != nil {
		return LoginOutput{}, err
	}

	return LoginOutput{
		AccessToken: tok,
		UserID:      u.ID,
		Email:       u.Email,
	}, nil
}

func (s *AuthService) VerifyEmail(verifyToken string) error {
	user, err := s.users.ByVerificationToken(verifyToken)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errInvalidToken
		}
		return err
	}

	if user.Status == domain.StatusActive {
		return errAlreadyActive
	}

	return s.users.Activate(user.ID)
}
