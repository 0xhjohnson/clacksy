package clacksy

import (
	"context"
	"strings"
	"time"
)

type Soundtest struct {
	SoundtestID int
	UserID      int
	User        *User
	// Location of the soundtest audio file
	URL string

	// Keyboard parts
	KeyboardID       int
	Keyboard         *Keyboard
	KeycapMaterialID int
	KeycapMaterial   *KeycapMaterial
	PlateMaterialID  int
	PlateMaterial    *PlateMaterial
	KeyswitchID      int
	Keyswitch        *Keyswitch

	CreatedAt  time.Time
	UpdatedAt  time.Time
	FeaturedOn time.Time
}

// SoundtestService represents a service for managing soundtests.
type SoundtestService interface {
	CreateSoundtest(ctx context.Context, soundtest *Soundtest) error

	// Retrieves a single soundtest by ID. Returns ENOTFOUND if soundtest does
	// not exist.
	FindSoundtestByID(ctx context.Context, soundtestID int) (*Soundtest, error)

	// Retrieves a list of soundtests based on a filter. Also returns a count
	// of total matching soundtests which can differ from the number of returned
	// soundtests if "Limit" field is set.
	FindSoundtests(ctx context.Context, filter SoundtestFilter) ([]*Soundtest, int, error)

	// These are primarily only used in tests.
	CreateKeyswitch(ctx context.Context, keyswitch *Keyswitch) error
	CreateKeyboard(ctx context.Context, keyboard *Keyboard) error
	CreateKeycapMaterial(ctx context.Context, keycapMaterial *KeycapMaterial) error
	CreatePlateMaterial(ctx context.Context, plateMaterial *PlateMaterial) error
}

func (s *Soundtest) Validate() error {
	switch {
	case strings.TrimSpace(s.URL) == "":
		return Errorf(EINVALID, "Soundtest URL is required.")
	case s.KeyboardID == 0:
		return Errorf(EINVALID, "Soundtest keyboard is required.")
	case s.KeycapMaterialID == 0:
		return Errorf(EINVALID, "Soundtest keycap material is required.")
	case s.PlateMaterialID == 0:
		return Errorf(EINVALID, "Soundtest plate material is required.")
	case s.KeyswitchID == 0:
		return Errorf(EINVALID, "Soundtest keyswitch is required.")
	default:
		return nil
	}
}

type SoundtestFilter struct {
	SoundtestID *int
	UserID      *int

	Offset int
	Limit  int
}

type Keyboard struct {
	KeyboardID int
	Name       string
}

type KeycapMaterial struct {
	KeycapMaterialID int
	Name             string
}

type PlateMaterial struct {
	PlateMaterialID int
	Name            string
}

type Keyswitch struct {
	KeyswitchID   int
	Name          string
	KeyswitchType *KeyswitchType
}

type KeyswitchType struct {
	KeyswitchTypeID int
	Name            string
}
