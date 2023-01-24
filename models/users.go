package models

import (
	"context"
	"errors"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID             uuid.UUID
	Email          string
	HashedPassword []byte
	Created        time.Time
	LastUpdated    time.Time
	Name           string
	Username       string
}

type UserModel struct {
	DB *pgxpool.Pool
}

func (m *UserModel) Insert(email, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	stmt := `INSERT INTO user_profile (email, hashed_password, created)
		VALUES ($1, $2, now())`

	_, err = m.DB.Exec(context.Background(), stmt, email, hashedPassword)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return ErrDuplicateEmail
			}
		}
		return err
	}

	return nil
}

func (m *UserModel) Authenticate(email, password string) (uuid.UUID, error) {
	var userID uuid.UUID
	var hashedPassword []byte

	stmt := `SELECT user_profile_id, hashed_password
		FROM user_profile
		WHERE email = $1`

	err := m.DB.QueryRow(context.Background(), stmt, email).Scan(&userID, &hashedPassword)
	if err != nil {
		if err == pgx.ErrNoRows {
			return userID, ErrInvalidCredentials
		}
		return userID, err
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return userID, ErrInvalidCredentials
		}
		return userID, err
	}

	return userID, nil
}

func (m *UserModel) Exists(id string) (bool, error) {
	var exists bool

	stmt := "SELECT EXISTS(SELECT true FROM user_profile WHERE user_profile_id = $1)"

	err := m.DB.QueryRow(context.Background(), stmt, id).Scan(&exists)

	return exists, err
}

type ProfileInfo struct {
	ID          uuid.UUID
	Email       string
	LastUpdated time.Time
	Name        string
	Username    string
}

func (m *UserModel) GetProfileInfo(userID string) (ProfileInfo, error) {
	var p ProfileInfo

	stmt := `SELECT
				user_profile_id,
				email,
				last_updated,
				COALESCE(name, ''),
				COALESCE(username, '')
			FROM user_profile
			WHERE user_profile_id = $1`

	err := m.DB.QueryRow(context.Background(), stmt, userID).Scan(&p.ID, &p.Email, &p.LastUpdated, &p.Name, &p.Username)
	if err != nil {
		return p, err
	}

	return p, nil
}

func (m *UserModel) UpdateProfile(userID, email, name, username string) error {
	stmt := `UPDATE user_profile
		SET email = $2, name = $3, username = NULLIF($4, ''), last_updated = now()
		WHERE user_profile_id = $1`

	_, err := m.DB.Exec(context.Background(), stmt, userID, email, name, username)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return ErrDuplicateUsername
			}
		}
		return err
	}

	return nil
}
