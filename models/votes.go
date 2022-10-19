package models

import (
	"context"
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

func (m *VoteModel) Upsert(soundtestID string, voteType int, userID string) error {
	stmt := `INSERT INTO vote (sound_test_id, vote_type, created_by)
	VALUES ($1, $2, $3)
	ON CONFLICT (sound_test_id, created_by)
	DO UPDATE SET vote_type = EXCLUDED.vote_type`

	_, err := m.DB.Exec(context.Background(), stmt, soundtestID, voteType, userID)
	if err != nil {
		return err
	}

	return nil
}
