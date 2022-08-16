package models

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
)

type SoundTest struct {
	ID               uuid.UUID
	URL              string
	Uploaded         time.Time
	LastUpdated      time.Time
	KeyboardID       uuid.UUID
	PlateMaterialID  uuid.UUID
	KeycapMaterialID uuid.UUID
	KeyswitchID      uuid.UUID
	CreatedBy        uuid.UUID
}

type SoundTestModel struct {
	DB *pgxpool.Pool
}

func (m *SoundTestModel) Insert(fileURL, keyboard, plateMaterial, keycapMaterial, keyswitch, userID string) error {
	stmt := `INSERT INTO sound_test (url, uploaded, keyboard_id, plate_material_id, keycap_material_id, keyswitch_id, created_by)
		VALUES ($1, now(), $2, $3, $4, $5, $6)`

	_, err := m.DB.Exec(context.Background(), stmt, fileURL, keyboard, plateMaterial, keycapMaterial, keyswitch, userID)
	if err != nil {
		return err
	}

	return nil
}
