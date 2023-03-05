package sqlite

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"github.com/0xhjohnson/clacksy"
)

var _ clacksy.SoundtestService = (*SoundtestService)(nil)

type SoundtestService struct {
	db *DB
}

func NewSoundtestService(db *DB) *SoundtestService {
	return &SoundtestService{db: db}
}

func (s *SoundtestService) CreateSoundtest(ctx context.Context, soundtest *clacksy.Soundtest) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := createSoundtest(ctx, tx, soundtest); err != nil {
		return err
	}

	err = attachSoundtestAssociations(ctx, tx, soundtest)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *SoundtestService) UpsertVote(ctx context.Context, vote *clacksy.SoundtestVote) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := upsertVote(ctx, tx, vote); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *SoundtestService) FindSoundtestByID(ctx context.Context, soundtestID int) (*clacksy.Soundtest, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	soundtest, err := findSoundtestByID(ctx, tx, soundtestID)
	if err != nil {
		return nil, err
	} else if err := attachSoundtestAssociations(ctx, tx, soundtest); err != nil {
		return nil, err
	}

	return soundtest, nil
}

func (s *SoundtestService) FindSoundtests(ctx context.Context, filter clacksy.SoundtestFilter) ([]*clacksy.Soundtest, int, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback()

	soundtests, n, err := findSoundtests(ctx, tx, filter)
	if err != nil {
		return soundtests, n, err
	}

	for _, soundtest := range soundtests {
		if err := attachSoundtestAssociations(ctx, tx, soundtest); err != nil {
			return soundtests, n, err
		}
	}
	return soundtests, n, nil
}

func (s *SoundtestService) FindKeyboards(ctx context.Context, filter clacksy.KeyboardFilter) ([]*clacksy.Keyboard, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	keyboards, err := findKeyboards(ctx, tx, filter)
	if err != nil {
		return keyboards, err
	}

	return keyboards, nil
}

func (s *SoundtestService) FindKeyswitches(ctx context.Context, filter clacksy.KeyswitchFilter) ([]*clacksy.Keyswitch, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	keyswitches, err := findKeyswitches(ctx, tx, filter)
	if err != nil {
		return keyswitches, err
	}

	return keyswitches, nil
}

func (s *SoundtestService) FindPlateMaterials(ctx context.Context, filter clacksy.PlateMaterialFilter) ([]*clacksy.PlateMaterial, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	plateMaterials, err := findPlateMaterials(ctx, tx, filter)
	if err != nil {
		return plateMaterials, err
	}

	return plateMaterials, nil
}

func (s *SoundtestService) FindKeycapMaterials(ctx context.Context, filter clacksy.KeycapMaterialFilter) ([]*clacksy.KeycapMaterial, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	keycapMaterials, err := findKeycapMaterials(ctx, tx, filter)
	if err != nil {
		return keycapMaterials, err
	}

	return keycapMaterials, nil
}

func (s *SoundtestService) CreateKeyboard(ctx context.Context, keyboard *clacksy.Keyboard) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := createKeyboard(ctx, tx, keyboard); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *SoundtestService) CreateKeyswitch(ctx context.Context, keyswitch *clacksy.Keyswitch) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := createKeyswitch(ctx, tx, keyswitch); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *SoundtestService) CreateKeycapMaterial(ctx context.Context, keycapMaterial *clacksy.KeycapMaterial) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := createKeycapMaterial(ctx, tx, keycapMaterial); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *SoundtestService) CreatePlateMaterial(ctx context.Context, plateMaterial *clacksy.PlateMaterial) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := createPlateMaterial(ctx, tx, plateMaterial); err != nil {
		return err
	}

	return tx.Commit()
}

func findSoundtestByID(ctx context.Context, tx *Tx, soundtestID int) (*clacksy.Soundtest, error) {
	soundtests, _, err := findSoundtests(ctx, tx, clacksy.SoundtestFilter{SoundtestID: &soundtestID})
	if err != nil {
		return nil, err
	} else if len(soundtests) == 0 {
		return nil, &clacksy.Error{Code: clacksy.ENOTFOUND, Message: "Soundtest not found."}
	}

	return soundtests[0], nil
}

func createSoundtest(ctx context.Context, tx *Tx, soundtest *clacksy.Soundtest) error {
	userID := clacksy.UserIDFromContext(ctx)
	if userID == 0 {
		return clacksy.Errorf(clacksy.EUNAUTHORIZED, "You must be logged in to create a soundtest.")
	}
	soundtest.UserID = clacksy.UserIDFromContext(ctx)

	soundtest.CreatedAt = tx.now
	soundtest.UpdatedAt = soundtest.CreatedAt

	if err := soundtest.Validate(); err != nil {
		return err
	} else if _, err := findUserByID(ctx, tx, soundtest.UserID); err != nil {
		return err
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO soundtest (
			user_id,
			keyboard_id,
			plate_material_id,
			keycap_material_id,
			keyswitch_id,
			url,
			created_at,
			updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`,
		soundtest.UserID,
		soundtest.KeyboardID,
		soundtest.PlateMaterialID,
		soundtest.KeycapMaterialID,
		soundtest.KeyswitchID,
		soundtest.URL,
		(*NullTime)(&soundtest.CreatedAt),
		(*NullTime)(&soundtest.UpdatedAt),
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	soundtest.SoundtestID = int(id)

	return nil
}

func upsertVote(ctx context.Context, tx *Tx, vote *clacksy.SoundtestVote) error {
	userID := clacksy.UserIDFromContext(ctx)
	if userID == 0 {
		return clacksy.Errorf(clacksy.EUNAUTHORIZED, "You must be logged in to vote.")
	}
	vote.UserID = clacksy.UserIDFromContext(ctx)

	vote.UpdatedAt = tx.now

	if err := vote.Validate(); err != nil {
		return err
	} else if _, err := findUserByID(ctx, tx, vote.UserID); err != nil {
		return err
	}

	_, err := tx.ExecContext(ctx, `
		INSERT INTO vote (soundtest_id, user_id, vote_type, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT (soundtest_id, user_id)
		DO UPDATE SET vote_type = excluded.vote_type
	`,
		vote.SoundtestID,
		vote.UserID,
		vote.VoteType,
		(*NullTime)(&vote.UpdatedAt),
	)
	if err != nil {
		return err
	}

	return nil
}

func findSoundtests(ctx context.Context, tx *Tx, filter clacksy.SoundtestFilter) (_ []*clacksy.Soundtest, n int, err error) {
	// Build WHERE clause. Each part of the WHERE clause is AND-ed together.
	// Values are appended to an arg list to avoid SQL injection.
	where, args := []string{"1 = 1"}, []interface{}{}
	if v := filter.SoundtestID; v != nil {
		where, args = append(where, "soundtest_id = ?"), append(args, *v)
	} else if v := filter.UserID; v != nil {
		where, args = append(where, "user_id = ?"), append(args, *v)
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT
			soundtest_id,
			user_id,
			keyboard_id,
			plate_material_id,
			keycap_material_id,
			keyswitch_id,
			url,
			featured_on,
			created_at,
			updated_at,
			COUNT(*) OVER()
		FROM soundtest
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY user_id ASC
		`+FormatLimitOffset(filter.Limit, filter.Offset),
		args...,
	)
	if err != nil {
		return nil, n, err
	}
	defer rows.Close()

	soundtests := make([]*clacksy.Soundtest, 0)
	for rows.Next() {
		var soundtest clacksy.Soundtest

		err := rows.Scan(
			&soundtest.SoundtestID,
			&soundtest.UserID,
			&soundtest.KeyboardID,
			&soundtest.PlateMaterialID,
			&soundtest.KeycapMaterialID,
			&soundtest.KeyswitchID,
			&soundtest.URL,
			(*NullTime)(&soundtest.FeaturedOn),
			(*NullTime)(&soundtest.CreatedAt),
			(*NullTime)(&soundtest.UpdatedAt),
			&n,
		)
		if err != nil {
			return nil, n, err
		}

		soundtests = append(soundtests, &soundtest)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return soundtests, n, nil
}

func findKeyboards(ctx context.Context, tx *Tx, filter clacksy.KeyboardFilter) ([]*clacksy.Keyboard, error) {
	where, args := []string{"1 = 1"}, []interface{}{}
	keebFilter := filter.KeyboardID
	if keebFilter != nil {
		stmt := `keyboard_id = ? OR
				 keyboard_id IN (
					SELECT keyboard_id
					FROM keyboard
					WHERE keyboard_id != ?
					ORDER BY RANDOM()
					LIMIT 3
				 )`
		where, args = append(where, stmt), append(args, *keebFilter, *keebFilter)
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT
			keyboard_id,
			name
		FROM keyboard
		WHERE `+strings.Join(where, " AND "),
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	keyboards := make([]*clacksy.Keyboard, 0)
	for rows.Next() {
		var keyboard clacksy.Keyboard

		err := rows.Scan(&keyboard.KeyboardID, &keyboard.Name)
		if err != nil {
			return nil, err
		}

		keyboards = append(keyboards, &keyboard)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if keebFilter != nil {
		rand.Shuffle(len(keyboards), func(i, j int) {
			keyboards[i], keyboards[j] = keyboards[j], keyboards[i]
		})
	}

	return keyboards, nil
}

func findKeyswitches(ctx context.Context, tx *Tx, filter clacksy.KeyswitchFilter) ([]*clacksy.Keyswitch, error) {
	where, args := []string{"1 = 1"}, []interface{}{}
	switchFilter := filter.KeyswitchID
	if switchFilter != nil {
		stmt := `keyswitch_id = ? OR
				 keyswitch_id IN (
					SELECT keyswitch_id
					FROM keyswitch
					WHERE keyswitch_id != ?
					ORDER BY RANDOM()
					LIMIT 3
				 )`
		where, args = append(where, stmt), append(args, *switchFilter, *switchFilter)
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT
			k.keyswitch_id,
			k.name,
			k.keyswitch_type_id,
			kt.name keyswitch_type
		FROM keyswitch k
		RIGHT JOIN keyswitch_type kt USING (keyswitch_type_id)
		WHERE `+strings.Join(where, " AND "),
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	switches := make([]*clacksy.Keyswitch, 0)
	for rows.Next() {
		keyswitch := &clacksy.Keyswitch{
			KeyswitchType: &clacksy.KeyswitchType{},
		}

		err := rows.Scan(
			&keyswitch.KeyswitchID,
			&keyswitch.Name,
			&keyswitch.KeyswitchType.KeyswitchTypeID,
			&keyswitch.KeyswitchType.Name,
		)
		if err != nil {
			return nil, err
		}

		switches = append(switches, keyswitch)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if switchFilter != nil {
		rand.Shuffle(len(switches), func(i, j int) {
			switches[i], switches[j] = switches[j], switches[i]
		})
	}

	return switches, nil
}

func findPlateMaterials(ctx context.Context, tx *Tx, filter clacksy.PlateMaterialFilter) ([]*clacksy.PlateMaterial, error) {
	where, args := []string{"1 = 1"}, []interface{}{}
	plateMatFilter := filter.PlateMaterialID
	if plateMatFilter != nil {
		stmt := `plate_material_id = ? OR
				 plate_material_id IN (
					SELECT plate_material_id
					FROM plate_material
					WHERE plate_material_id != ?
					ORDER BY RANDOM()
					LIMIT 3
				 )`
		where, args = append(where, stmt), append(args, *plateMatFilter, *plateMatFilter)
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT
			plate_material_id,
			name
		FROM plate_material
		WHERE `+strings.Join(where, " AND "),
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	plateMaterials := make([]*clacksy.PlateMaterial, 0)
	for rows.Next() {
		var plateMaterial clacksy.PlateMaterial

		err := rows.Scan(&plateMaterial.PlateMaterialID, &plateMaterial.Name)
		if err != nil {
			return nil, err
		}

		plateMaterials = append(plateMaterials, &plateMaterial)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if plateMatFilter != nil {
		rand.Shuffle(len(plateMaterials), func(i, j int) {
			plateMaterials[i], plateMaterials[j] = plateMaterials[j], plateMaterials[i]
		})
	}

	return plateMaterials, nil
}

func findKeycapMaterials(ctx context.Context, tx *Tx, filter clacksy.KeycapMaterialFilter) ([]*clacksy.KeycapMaterial, error) {
	where, args := []string{"1 = 1"}, []interface{}{}
	keycapMatFilter := filter.KeycapMaterialID
	if keycapMatFilter != nil {
		stmt := `keycap_material_id = ? OR
				 keycap_material_id IN (
					SELECT keycap_material_id
					FROM keycap_material
					WHERE keycap_material_id != ?
					ORDER BY RANDOM()
					LIMIT 3
				 )`
		where, args = append(where, stmt), append(args, *keycapMatFilter, *keycapMatFilter)
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT
			keycap_material_id,
			name
		FROM keycap_material
		WHERE `+strings.Join(where, " AND "),
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	keycapMaterials := make([]*clacksy.KeycapMaterial, 0)
	for rows.Next() {
		var keycapMaterial clacksy.KeycapMaterial

		err := rows.Scan(&keycapMaterial.KeycapMaterialID, &keycapMaterial.Name)
		if err != nil {
			return nil, err
		}

		keycapMaterials = append(keycapMaterials, &keycapMaterial)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if keycapMatFilter != nil {
		rand.Shuffle(len(keycapMaterials), func(i, j int) {
			keycapMaterials[i], keycapMaterials[j] = keycapMaterials[j], keycapMaterials[i]
		})
	}

	return keycapMaterials, nil
}

func createKeyboard(ctx context.Context, tx *Tx, keyboard *clacksy.Keyboard) error {
	result, err := tx.ExecContext(ctx, `
		INSERT INTO keyboard (name)
		VALUES (?)
	`,
		keyboard.Name,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	keyboard.KeyboardID = int(id)

	return nil
}

func createKeyswitch(ctx context.Context, tx *Tx, keyswitch *clacksy.Keyswitch) error {
	result, err := tx.ExecContext(ctx, `
		INSERT INTO keyswitch_type (name)
		VALUES (?)
	`,
		keyswitch.KeyswitchType.Name,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	keyswitch.KeyswitchType.KeyswitchTypeID = int(id)

	result, err = tx.ExecContext(ctx, `
		INSERT INTO keyswitch (name, keyswitch_type_id)
		VALUES (?, ?)
	`,
		keyswitch.Name,
		keyswitch.KeyswitchType.KeyswitchTypeID,
	)
	if err != nil {
		return err
	}

	id, err = result.LastInsertId()
	if err != nil {
		return err
	}
	keyswitch.KeyswitchID = int(id)

	return nil
}

func createKeycapMaterial(ctx context.Context, tx *Tx, keycapMaterial *clacksy.KeycapMaterial) error {
	result, err := tx.ExecContext(ctx, `
		INSERT INTO keycap_material (name)
		VALUES (?)
	`,
		keycapMaterial.Name,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	keycapMaterial.KeycapMaterialID = int(id)

	return nil
}

func createPlateMaterial(ctx context.Context, tx *Tx, plateMaterial *clacksy.PlateMaterial) error {
	result, err := tx.ExecContext(ctx, `
		INSERT INTO plate_material (name)
		VALUES (?)
	`,
		plateMaterial.Name,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	plateMaterial.PlateMaterialID = int(id)

	return nil
}

func attachSoundtestAssociations(ctx context.Context, tx *Tx, soundtest *clacksy.Soundtest) (err error) {
	soundtest.User, err = findUserByID(ctx, tx, soundtest.UserID)
	if err != nil {
		return fmt.Errorf("attach soundtest user: %w", err)
	}

	return nil
}
