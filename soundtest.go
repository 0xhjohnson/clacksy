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
	UserVote    int
	TotalVotes  int
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
	UpsertVote(ctx context.Context, vote *SoundtestVote) error

	// Retrieves a single soundtest by ID. Returns ENOTFOUND if soundtest does
	// not exist.
	FindSoundtestByID(ctx context.Context, soundtestID int) (*Soundtest, error)

	// Retrieves a list of soundtests based on a filter. Also returns a count
	// of total matching soundtests which can differ from the number of returned
	// soundtests if "Limit" field is set.
	FindSoundtests(ctx context.Context, filter SoundtestFilter) ([]*Soundtest, int, error)

	// Retrieves a list of keyboards based on a filter. The filter is used to
	// retrieve the correct keyboard (filter.KeyboardID) along with 3 others.
	FindKeyboards(ctx context.Context, filter KeyboardFilter) ([]*Keyboard, error)

	// Retrieves a list of keyswitches based on a filter. The filter is used to
	// retrieve the correct keyswitch (filter.KeyswitchID) along with 3 others.
	FindKeyswitches(ctx context.Context, filter KeyswitchFilter) ([]*Keyswitch, error)

	// Retrieves a list of plate materials based on a filter. The filter is used to
	// retrieve the correct plate material (filter.PlateMaterialID) along with 3 others.
	FindPlateMaterials(ctx context.Context, filter PlateMaterialFilter) ([]*PlateMaterial, error)

	// Retrieves a list of keycap materials based on a filter. The filter is used to
	// retrieve the correct keycap material (filter.KeycapMaterialID) along with 3 others.
	FindKeycapMaterials(ctx context.Context, filter KeycapMaterialFilter) ([]*KeycapMaterial, error)

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

func (sv *SoundtestVote) Validate() error {
	switch {
	case sv.SoundtestID == 0:
		return Errorf(EINVALID, "Soundtest ID is required.")
	case sv.VoteType < -1 || sv.VoteType > 1:
		return Errorf(EINVALID, "Soundtest vote type must be -1, 0, or 1.")
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

type SoundtestVote struct {
	SoundtestID int
	UserID      int
	VoteType    int
	UpdatedAt   time.Time
}

type KeyboardFilter struct {
	KeyboardID *int
}

type KeyswitchFilter struct {
	KeyswitchID *int
}

type KeycapMaterialFilter struct {
	KeycapMaterialID *int
}

type PlateMaterialFilter struct {
	PlateMaterialID *int
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
