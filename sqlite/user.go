package sqlite

import (
	"context"
	"database/sql"
	"errors"
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

func (s *UserService) FindUserByID(ctx context.Context, id int) (*clacksy.User, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	user, err := findUserByID(ctx, tx, id)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Authenticate(ctx context.Context, user *clacksy.User, password string) (*clacksy.User, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	return authenticate(ctx, tx, user, password)
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

func findUserByID(ctx context.Context, tx *Tx, id int) (*clacksy.User, error) {
	var user clacksy.User
	var name sql.NullString

	err := tx.QueryRowContext(ctx, `
		SELECT
			user_id,
			name,
			email,
			hashed_password,
			created_at,
			updated_at
		FROM user
		WHERE user_id = ?
	`, id).Scan(
		&user.UserID,
		&name,
		&user.Email,
		&user.HashedPassword,
		(*NullTime)(&user.CreatedAt),
		(*NullTime)(&user.UpdatedAt),
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &clacksy.Error{Code: clacksy.ENOTFOUND, Message: "User not found."}
		}
		return nil, err
	}

	if name.Valid {
		user.Name = name.String
	}

	return &user, err
}

func authenticate(ctx context.Context, tx *Tx, user *clacksy.User, password string) (*clacksy.User, error) {
	if err := user.Validate(); err != nil {
		return nil, err
	}

	if err := validatePassword(password); err != nil {
		return nil, err
	}

	var u clacksy.User
	var name sql.NullString

	err := tx.QueryRowContext(ctx, `
		SELECT
			user_id,
			name,
			email,
			hashed_password,
			created_at,
			updated_at
		FROM user
		WHERE email = ?
	`, user.Email).Scan(
		&u.UserID,
		&name,
		&u.Email,
		&u.HashedPassword,
		(*NullTime)(&u.CreatedAt),
		(*NullTime)(&u.UpdatedAt),
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &clacksy.Error{Code: clacksy.ENOTFOUND, Message: "User not found."}
		}
		return nil, err
	}

	if name.Valid {
		u.Name = name.String
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.HashedPassword), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, &clacksy.Error{Code: clacksy.EINVALID, Message: "User credentials invalid."}
		}
		return nil, err
	}

	*user = u

	return user, err
}

func createUser(ctx context.Context, tx *Tx, user *clacksy.User, password string) error {
	user.CreatedAt = tx.now
	user.UpdatedAt = user.CreatedAt

	if err := user.Validate(); err != nil {
		return err
	}

	if err := validatePassword(password); err != nil {
		return err
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
		(*NullTime)(&user.CreatedAt),
		(*NullTime)(&user.UpdatedAt),
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

func validatePassword(password string) error {
	if strings.TrimSpace(password) == "" {
		return clacksy.Errorf(clacksy.EINVALID, "User password is required.")
	} else if utf8.RuneCountInString(password) <= 8 {
		return clacksy.Errorf(clacksy.EINVALID, "User password must be at least 8 characters.")
	}

	return nil
}
