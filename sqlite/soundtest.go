package sqlite

import (
	"context"
	"fmt"
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
