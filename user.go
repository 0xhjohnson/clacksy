package clacksy

import (
	"context"
	"strings"
	"time"
)

type User struct {
	UserID         int
	Name           string
	Email          string
	HashedPassword string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Validate returns an error if the user contains invalid fields.
// This only performs basic validation, more complex validations are done within
// subpackages such as sqlite/user.
func (u *User) Validate() error {
	switch {
	case strings.TrimSpace(u.Email) == "":
		return Errorf(EINVALID, "User email is required.")
	default:
		return nil
	}
}

// UserService represents a service for managing users.
type UserService interface {
	CreateUser(ctx context.Context, user *User, password string) error

	// Updates a user object. Returns EUNAUTHORIZED if current user is not
	// the user that is being updated. Returns ENOTFOUND if user does not exist.
	UpdateUser(ctx context.Context, id int, upd UserUpdate) (*User, error)

	// Authenticates a user. Returns ENOTFOUND if user does not exist.
	// Returns EINVALID if credentials are invalid.
	Authenticate(ctx context.Context, user *User, password string) (*User, error)

	// Retrieves a user by ID along with their associated auth objects.
	// Returns ENOTFOUND if user does not exist.
	FindUserByID(ctx context.Context, id int) (*User, error)
}

type UserUpdate struct {
	Name  *string
	Email *string
}
