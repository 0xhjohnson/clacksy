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
	FeaturedOn       time.Time
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

func (m *SoundTestModel) GetDaily() (SoundTest, error) {
	var st SoundTest

	stmt := `SELECT
		  sound_test_id,
		  url,
		  uploaded,
		  last_updated,
		  keyboard_id,
		  plate_material_id,
		  keycap_material_id,
		  keyswitch_id,
		  created_by,
		  featured_on
		FROM sound_test
		WHERE featured_on IS NOT NULL
		ORDER BY featured_on DESC
		LIMIT 1`

	err := m.DB.QueryRow(context.Background(), stmt).Scan(&st.ID, &st.URL, &st.Uploaded, &st.LastUpdated, &st.KeyboardID, &st.PlateMaterialID, &st.KeycapMaterialID, &st.KeyswitchID, &st.CreatedBy, &st.FeaturedOn)
	if err != nil {
		return st, err
	}

	return st, nil
}

func (m *SoundTestModel) AddPlay(soundtest uuid.UUID, userID, keyboard, plateMaterial, keycapMaterial, keyswitch string) error {
	stmt := `INSERT INTO sound_test_play (sound_test_id, created_by, submitted, keyboard_id, plate_material_id, keycap_material_id, keyswitch_id)
		VALUES($1, $2, now(), $3, $4, $5, $6)`

	_, err := m.DB.Exec(context.Background(), stmt, soundtest, userID, keyboard, plateMaterial, keycapMaterial, keyswitch)
	if err != nil {
		return err
	}

	return nil
}

type SoundTestPlay struct {
	SoundTestID           uuid.UUID
	URL                   string
	Submitted             time.Time
	CreatedBy             string
	Keyboard              string
	CorrectKeyboard       string
	PlateMaterial         string
	CorrectPlateMaterial  string
	KeycapMaterial        string
	CorrectKeycapMaterial string
	Keyswitch             string
	CorrectKeyswitch      string
}

func (m *SoundTestModel) GetPlay(userID string) (SoundTestPlay, error) {
	var p SoundTestPlay

	stmt := `SELECT
		stp.sound_test_id,
		st.url,
		stp.submitted,
		COALESCE(up.username, 'anonymous') created_by,
		k.name keyboard,
		ck.name correct_keyboard,
		pm.name plate_material,
		cpm.name correct_plate_material,
		km.name keycap_material,
		ckm.name correct_keycap_material,
		ks.name keyswitch,
		cks.name correct_keyswitch
	FROM sound_test_play stp
	JOIN sound_test st USING (sound_test_id)
	JOIN user_profile up ON st.created_by = up.user_profile_id
	JOIN keyboard k ON stp.keyboard_id = k.keyboard_id
	JOIN keyboard ck ON st.keyboard_id = ck.keyboard_id
	JOIN plate_material pm ON stp.plate_material_id = pm.plate_material_id
	JOIN plate_material cpm ON st.plate_material_id = cpm.plate_material_id
	JOIN keycap_material km ON stp.keycap_material_id = km.keycap_material_id
	JOIN keycap_material ckm ON st.keycap_material_id = ckm.keycap_material_id
	JOIN keyswitch ks ON stp.keyswitch_id = ks.keyswitch_id
	JOIN keyswitch cks ON st.keyswitch_id = cks.keyswitch_id
	WHERE
		stp.sound_test_id = (
			SELECT sound_test_id
			FROM sound_test
			WHERE featured_on IS NOT NULL
			ORDER BY featured_on DESC
			LIMIT 1
		)
		AND stp.created_by = $1`

	err := m.DB.QueryRow(context.Background(), stmt, userID).Scan(&p.SoundTestID, &p.URL, &p.Submitted, &p.CreatedBy, &p.Keyboard, &p.CorrectKeyboard, &p.PlateMaterial, &p.CorrectPlateMaterial, &p.KeycapMaterial, &p.CorrectKeycapMaterial, &p.Keyswitch, &p.CorrectKeyswitch)
	if err != nil {
		return p, err
	}

	return p, nil
}
