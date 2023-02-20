package sqlite

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/0xhjohnson/clacksy"
	"golang.org/x/crypto/bcrypt"
)

var _ clacksy.UserService = (*UserService)(nil)

type UserService struct {
	db *DB
}

func NewUserService(db *DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) CreateUser(ctx context.Context, user *clacksy.User, password string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := createUser(ctx, tx, user, password); err != nil {
		return err
	}
	return tx.Commit()
}

func createUser(ctx context.Context, tx *Tx, user *clacksy.User, password string) error {
	user.CreatedAt = tx.now
	user.UpdatedAt = user.CreatedAt

	if err := user.Validate(); err != nil {
		return err
	}

	// Perform password validations prior to hashing via bcrypt.
	if strings.TrimSpace(password) == "" {
		return clacksy.Errorf(clacksy.EINVALID, "User password is required.")
	} else if utf8.RuneCountInString(password) <= 8 {
		return clacksy.Errorf(clacksy.EINVALID, "User password must be at least 8 characters.")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}
	user.HashedPassword = string(hashedPassword[:])

	result, err := tx.ExecContext(ctx, `
		INSERT INTO user (
			name,
			email,
			hashed_password,
			created_at,
			updated_at
		)
		VALUES (?, ?, ?, ?, ?)
	`,
		user.Name,
		user.Email,
		user.HashedPassword,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	user.UserID = int(id)

	return nil
}
