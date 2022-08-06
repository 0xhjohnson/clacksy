package models

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	pgtypeuuid "github.com/jackc/pgtype/ext/gofrs-uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID              pgtypeuuid.UUID
	Email           string
	HashedPassword  []byte
	Created         time.Time
	LastUpdated     time.Time
	Name            string
	Username        string
	GenderPronounID pgtypeuuid.UUID
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
