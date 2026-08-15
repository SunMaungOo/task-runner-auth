package test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/SunMaungOo/task-runner-auth/internal/model"
	"github.com/SunMaungOo/task-runner-auth/internal/repo"
	"github.com/SunMaungOo/task-runner-auth/internal/service"
)

type mockRepo struct {
	users     map[string]model.User
	createErr error
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		users: make(map[string]model.User),
	}
}

func (mock *mockRepo) Create(context context.Context, email string, passwordHash string) (model.User, error) {

	if mock.createErr != nil {
		return model.User{}, mock.createErr
	}

	if _, exists := mock.users[email]; exists {
		return model.User{}, repo.ErrorDuplicateEmail
	}

	user := model.User{
		ID:           "mock-id-" + email,
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
	}

	mock.users[email] = user

	return user, nil
}

func (mock *mockRepo) GetByEmail(context context.Context, email string) (model.User, error) {

	user, ok := mock.users[email]

	if !ok {
		return model.User{}, repo.ErrorNotFound
	}

	return user, nil
}

func (*mockRepo) Ping(context context.Context) error {
	return nil
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func newTestService(userRepo repo.UserRepository) *service.AuthService {
	return service.New(userRepo, []byte("mock-secret"), time.Hour, testLogger())
}

func TestAuthService_Register(t *testing.T) {

	t.Run("valid registration", func(t *testing.T) {

		svc := newTestService(newMockRepo())

		user, err := svc.Register(context.Background(), "test@example.com", "mockpassword")

		if err != nil {
			t.Fatalf("Register() unexpected error :%v", err)
		}

		if user.Email != "test@example.com" {
			t.Errorf("got email %q, want normalized lowercase", user.Email)
		}

		if user.PasswordHash == "mockpassword" || user.PasswordHash == "" {
			t.Errorf("password was not hashed : %q", user.PasswordHash)
		}

	})

	t.Run("duplicate email registration", func(t *testing.T) {

		svc := newTestService(newMockRepo())

		_, err := svc.Register(context.Background(), "test@example.com", "mockpassword")

		_, err = svc.Register(context.Background(), "test@example.com", "mockpassword")

		if !errors.Is(err, service.ErrorEmailTaken) {
			t.Errorf("got err %v, want %v", err, service.ErrorEmailTaken)
		}

	})

	invalidEmails := []string{"", "not-an-email", "missing.com"}

	for _, email := range invalidEmails {

		t.Run("rejects invalid email", func(t *testing.T) {
			svc := newTestService(newMockRepo())

			_, err := svc.Register(context.Background(), email, "password1")

			if !errors.Is(err, service.ErrorInvalidEmail) {
				t.Errorf("Register(%q) error = %v, want %v", email, err, service.ErrorInvalidEmail)
			}
		})
	}

	t.Run("reject short email", func(t *testing.T) {

		password := "short"

		svc := newTestService(newMockRepo())

		_, err := svc.Register(context.Background(), "test@example.com", password)

		if !errors.Is(err, service.ErrorWeakPassword) {
			t.Errorf("Register() with password %q error = %v, want %v", password, err, service.ErrorWeakPassword)
		}
	})

}

func TestAuthService_Login(t *testing.T) {

	t.Run("correct credential", func(t *testing.T) {

		svc := newTestService(newMockRepo())

		email := "test@example.com"

		password := "password"

		if _, err := svc.Register(context.Background(), email, password); err != nil {
			t.Fatalf("Register() unexpected error:%v", err)
		}

		token, err := svc.Login(context.Background(), email, password)

		if err != nil {
			t.Fatalf("Login() unexpected error:%v", err)
		}

		if token == "" {
			t.Fatalf("Login() return an empty token")
		}

		userId, err := svc.ValidateToken(token)

		if err != nil {
			t.Fatalf("ValidateToken() unexpected error: %v", err)
		}

		if userId == "" {
			t.Error("ValidateToken() returned an empty user ID")
		}

	})

	t.Run("wrong password rejected", func(t *testing.T) {

		svc := newTestService(newMockRepo())

		email := "test@example.com"

		correctPassword := "correctPassword"

		wrongPassword := "wrongPassword"

		svc.Register(context.Background(), email, correctPassword)

		_, err := svc.Login(context.Background(), email, wrongPassword)

		if !errors.Is(err, service.ErrorInvalidCredential) {

			t.Errorf("got err %v, want %v", err, service.ErrorInvalidCredential)

		}

	})

	t.Run("unknown email and password rejected", func(t *testing.T) {

		svc := newTestService(newMockRepo())

		email := "test@example.com"

		password := "password"

		_, err := svc.Login(context.Background(), email, password)

		if !errors.Is(err, service.ErrorInvalidCredential) {

			t.Errorf("got err %v, want %v", err, service.ErrorInvalidCredential)

		}

	})
}

func TestAuthService_ValidateToken(t *testing.T) {

	svc := newTestService(newMockRepo())

	_, err := svc.ValidateToken("not-a-real-token")

	if !errors.Is(err, service.ErrorInvalidToken) {
		t.Errorf("got err %v, want %v", err, service.ErrorInvalidToken)
	}

}
