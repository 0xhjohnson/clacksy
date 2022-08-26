package models

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/sync/errgroup"
)

type Keyboard struct {
	ID   uuid.UUID
	Name string
}

type Keyswitch struct {
	ID                uuid.UUID
	Name              string
	KeyswitchTypeID   uuid.UUID
	KeyswitchTypeName string
}

type PlateMaterial struct {
	ID   uuid.UUID
	Name string
}

type KeycapMaterial struct {
	ID   uuid.UUID
	Name string
}

type PartsModel struct {
	DB *pgxpool.Pool
}

type AllParts struct {
	Keyboards       []Keyboard
	Switches        []Keyswitch
	PlateMaterials  []PlateMaterial
	KeycapMaterials []KeycapMaterial
}

func (m *PartsModel) GetAll() (AllParts, error) {
	var ap AllParts
	g, _ := errgroup.WithContext(context.Background())

	g.Go(func() error {
		keyboards, err := m.GetKeyboards()
		if err == nil {
			ap.Keyboards = keyboards
		}
		return err
	})

	g.Go(func() error {
		switches, err := m.GetSwitches()
		if err == nil {
			ap.Switches = switches
		}
		return err
	})

	g.Go(func() error {
		plateMaterials, err := m.GetPlateMaterials()
		if err == nil {
			ap.PlateMaterials = plateMaterials
		}
		return err
	})

	g.Go(func() error {
		keycapMaterials, err := m.GetKeycapMaterials()
		if err == nil {
			ap.KeycapMaterials = keycapMaterials
		}
		return err
	})

	err := g.Wait()
	if err != nil {
		return ap, nil
	}

	return ap, nil
}

func (m *PartsModel) GetKeyboards() ([]Keyboard, error) {
	var keebs []Keyboard

	stmt := `SELECT keyboard_id, name
		FROM keyboard
		ORDER BY name`

	rows, err := m.DB.Query(context.Background(), stmt)
	if err != nil {
		return keebs, err
	}
	defer rows.Close()

	for rows.Next() {
		var k Keyboard

		err := rows.Scan(&k.ID, &k.Name)
		if err != nil {
			return keebs, err
		}

		keebs = append(keebs, k)
	}

	return keebs, nil
}

func (m *PartsModel) GetSwitches() ([]Keyswitch, error) {
	var switches []Keyswitch

	stmt := `SELECT
			k.keyswitch_id,
			k.name,
			kt.keyswitch_type_id,
			kt.name as keyswitch_type_name
		FROM keyswitch k
		JOIN keyswitch_type kt using (keyswitch_type_id)
		ORDER BY k.name`

	rows, err := m.DB.Query(context.Background(), stmt)
	if err != nil {
		return switches, err
	}
	defer rows.Close()

	for rows.Next() {
		var k Keyswitch

		err := rows.Scan(&k.ID, &k.Name, &k.KeyswitchTypeID, &k.KeyswitchTypeName)
		if err != nil {
			return switches, err
		}

		switches = append(switches, k)
	}

	return switches, nil
}

func (m *PartsModel) GetPlateMaterials() ([]PlateMaterial, error) {
	var plateMaterials []PlateMaterial

	stmt := `SELECT plate_material_id, name
		FROM plate_material
		ORDER BY name`

	rows, err := m.DB.Query(context.Background(), stmt)
	if err != nil {
		return plateMaterials, err
	}
	defer rows.Close()

	for rows.Next() {
		var p PlateMaterial

		err := rows.Scan(&p.ID, &p.Name)
		if err != nil {
			return plateMaterials, err
		}

		plateMaterials = append(plateMaterials, p)
	}

	return plateMaterials, nil
}

func (m *PartsModel) GetKeycapMaterials() ([]KeycapMaterial, error) {
	var keycapMaterials []KeycapMaterial

	stmt := `SELECT keycap_material_id, name
		FROM keycap_material
		ORDER BY name`

	rows, err := m.DB.Query(context.Background(), stmt)
	if err != nil {
		return keycapMaterials, err
	}
	defer rows.Close()

	for rows.Next() {
		var k KeycapMaterial

		err := rows.Scan(&k.ID, &k.Name)
		if err != nil {
			return keycapMaterials, err
		}

		keycapMaterials = append(keycapMaterials, k)
	}

	return keycapMaterials, nil
}
