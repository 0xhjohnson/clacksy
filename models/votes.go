package models

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Vote struct {
	ID        uuid.UUID
	VoteType  int
	Timestamp time.Time
	CreatedBy uuid.UUID
}

type VoteModel struct {
	DB *pgxpool.Pool
}
