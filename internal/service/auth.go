package service

import (
	"context"
	"fmt"
	"log/slog"
	"net/mail"
	"strings"
	"time"

	"errors"

	"github.com/SunMaungOo/task-runner-auth/internal/model"
	"github.com/SunMaungOo/task-runner-auth/internal/repo"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrorInvalidEmail      = errors.New("invalid email address")
	ErrorWeakPassword      = errors.New("weak password")
	ErrorEmailTaken        = errors.New("email taken")
	ErrorInvalidCredential = errors.New("invalid email or password")
	ErrorInvalidToken      = errors.New("invalid token")
)

type AuthService struct {
	userRepo  repo.UserRepository
	jwtSecret []byte
	tokenTTL  time.Duration
	logger    *slog.Logger
}

func New(userRepo repo.UserRepository, jwtSecret []byte, tokenTTL time.Duration, logger *slog.Logger) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		tokenTTL:  tokenTTL,
		logger:    logger,
	}
}

func (auth AuthService) Register(context context.Context, email string, password string) (model.User, error) {
	email = normalizeEmail(email)

	if _, err := mail.ParseAddress(email); err != nil {
		return model.User{}, ErrorInvalidEmail
	}

	if len(password) < 8 {
		return model.User{}, ErrorWeakPassword
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return model.User{}, fmt.Errorf("hash password %w", err)
	}

	user, err := auth.userRepo.Create(context, email, string(hash))

	if err != nil {

		if errors.Is(err, repo.ErrorDuplicateEmail) {
			return model.User{}, ErrorEmailTaken
		}

		auth.logger.ErrorContext(context, "create user failed", "error", err)

		return model.User{}, err
	}

	return user, nil
}

func (auth AuthService) Login(context context.Context, email string, password string) (string, error) {

	email = normalizeEmail(email)

	user, err := auth.userRepo.GetByEmail(context, email)

	if err != nil {

		if errors.Is(err, repo.ErrorNotFound) {
			return "", ErrorInvalidCredential
		}

		auth.logger.ErrorContext(context, "lookup user failed", "error", err)

		return "", err

	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {

		return "", ErrorInvalidCredential
	}

	return auth.issueToken(user)

}

func (auth AuthService) issueToken(user model.User) (string, error) {

	now := time.Now()

	claim := jwt.RegisteredClaims{
		Subject:   user.ID,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(auth.tokenTTL)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)

	return token.SignedString(auth.jwtSecret)
}

func (auth AuthService) ValidateToken(jwtToken string) (string, error) {

	token, err := jwt.Parse(jwtToken, func(providedToken *jwt.Token) (any, error) {

		if _, ok := providedToken.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", providedToken.Header["alg"])
		}

		return auth.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return "", ErrorInvalidToken
	}

	claim, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		return "", ErrorInvalidToken
	}

	return claim.GetSubject()

}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
