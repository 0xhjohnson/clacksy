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

type SoundTestVote struct {
	ID          uuid.UUID
	URL         string
	Uploaded    time.Time
	LastUpdated time.Time
	CreatedBy   string
	UserVote    int
	TotalVotes  int
	TotalTests  int
}

func (m *SoundTestModel) GetLatest(page int, perPage int, userID string) ([]SoundTestVote, error) {
	var soundtests []SoundTestVote

	stmt := `SELECT
		  st.sound_test_id,
		  st.url,
		  st.uploaded,
		  st.last_updated,
		  COALESCE(up.username, 'anonymous'),
		  COALESCE(
		    (SELECT vote_type
		    FROM vote
		    WHERE
		      vote.sound_test_id = st.sound_test_id
		      AND vote.created_by = $3
		    ), 0) as user_vote,
		  (SELECT COALESCE(SUM(vote_type), 0)
		  FROM vote
		  WHERE vote.sound_test_id = st.sound_test_id) as total_votes,
		  count(*) over() as total_tests
		FROM sound_test st
		JOIN user_profile up ON up.user_profile_id = st.created_by
		ORDER BY st.uploaded DESC
		OFFSET $1 * 10
		FETCH NEXT $2 ROWS ONLY`

	rows, err := m.DB.Query(context.Background(), stmt, page, perPage, userID)
	if err != nil {
		return soundtests, err
	}
	defer rows.Close()

	for rows.Next() {
		var st SoundTestVote

		err := rows.Scan(&st.ID, &st.URL, &st.Uploaded, &st.LastUpdated, &st.CreatedBy, &st.UserVote, &st.TotalVotes, &st.TotalTests)
		if err != nil {
			return soundtests, err
		}

		soundtests = append(soundtests, st)
	}

	return soundtests, nil
}
