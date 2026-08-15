package repo

import (
	"context"
	"errors"

	"github.com/SunMaungOo/task-runner-auth/internal/model"
)

var (
	ErrorDuplicateEmail = errors.New("email already registered")
	ErrorNotFound       = errors.New("user not found")
)

type UserRepository interface {
	Create(context context.Context, email string, passwordHash string) (model.User, error)

	GetByEmail(context context.Context, email string) (model.User, error)

	Ping(context context.Context) error
}
