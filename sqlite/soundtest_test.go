package sqlite_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/0xhjohnson/clacksy"
	"github.com/0xhjohnson/clacksy/sqlite"
	"github.com/google/go-cmp/cmp"
)

func TestSoundtestService_CreateSoundtest(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		user             *clacksy.User
		url              string
		keyboard         *clacksy.Keyboard
		keyswitch        *clacksy.Keyswitch
		keycapMaterial   *clacksy.KeycapMaterial
		plateMaterial    *clacksy.PlateMaterial
		wantErrorCode    string
		wantErrorMessage string
	}{
		"OK": {
			user:             &clacksy.User{Name: "john", Email: "john@gmail.com"},
			url:              "/soundtests/sonnet.mp4",
			keyboard:         &clacksy.Keyboard{Name: "mode sonnet"},
			keyswitch:        &clacksy.Keyswitch{Name: "boba lt", KeyswitchType: &clacksy.KeyswitchType{Name: "linear"}},
			keycapMaterial:   &clacksy.KeycapMaterial{Name: "abs"},
			plateMaterial:    &clacksy.PlateMaterial{Name: "pom"},
			wantErrorCode:    "",
			wantErrorMessage: "",
		},
		"ErrURLRequired": {
			user:             &clacksy.User{Name: "john", Email: "john@gmail.com"},
			url:              "",
			keyboard:         &clacksy.Keyboard{Name: "mode sonnet"},
			keyswitch:        &clacksy.Keyswitch{Name: "boba lt", KeyswitchType: &clacksy.KeyswitchType{Name: "linear"}},
			keycapMaterial:   &clacksy.KeycapMaterial{Name: "abs"},
			plateMaterial:    &clacksy.PlateMaterial{Name: "pom"},
			wantErrorCode:    clacksy.EINVALID,
			wantErrorMessage: "Soundtest URL is required.",
		},
		"ErrKeyboardRequired": {
			user:             &clacksy.User{Name: "john", Email: "john@gmail.com"},
			url:              "/soundtests/sonnet.mp4",
			keyboard:         nil,
			keyswitch:        &clacksy.Keyswitch{Name: "boba lt", KeyswitchType: &clacksy.KeyswitchType{Name: "linear"}},
			keycapMaterial:   &clacksy.KeycapMaterial{Name: "abs"},
			plateMaterial:    &clacksy.PlateMaterial{Name: "pom"},
			wantErrorCode:    clacksy.EINVALID,
			wantErrorMessage: "Soundtest keyboard is required.",
		},
		"ErrKeyswitchRequired": {
			user:             &clacksy.User{Name: "john", Email: "john@gmail.com"},
			url:              "/soundtests/sonnet.mp4",
			keyboard:         &clacksy.Keyboard{Name: "mode sonnet"},
			keyswitch:        nil,
			keycapMaterial:   &clacksy.KeycapMaterial{Name: "abs"},
			plateMaterial:    &clacksy.PlateMaterial{Name: "pom"},
			wantErrorCode:    clacksy.EINVALID,
			wantErrorMessage: "Soundtest keyswitch is required.",
		},
		"ErrKeycapMaterialRequired": {
			user:             &clacksy.User{Name: "john", Email: "john@gmail.com"},
			url:              "/soundtests/sonnet.mp4",
			keyboard:         &clacksy.Keyboard{Name: "mode sonnet"},
			keyswitch:        &clacksy.Keyswitch{Name: "boba lt", KeyswitchType: &clacksy.KeyswitchType{Name: "linear"}},
			keycapMaterial:   nil,
			plateMaterial:    &clacksy.PlateMaterial{Name: "pom"},
			wantErrorCode:    clacksy.EINVALID,
			wantErrorMessage: "Soundtest keycap material is required.",
		},
		"ErrPlateMaterialRequired": {
			user:             &clacksy.User{Name: "john", Email: "john@gmail.com"},
			url:              "/soundtests/sonnet.mp4",
			keyboard:         &clacksy.Keyboard{Name: "mode sonnet"},
			keyswitch:        &clacksy.Keyswitch{Name: "boba lt", KeyswitchType: &clacksy.KeyswitchType{Name: "linear"}},
			keycapMaterial:   &clacksy.KeycapMaterial{Name: "abs"},
			plateMaterial:    nil,
			wantErrorCode:    clacksy.EINVALID,
			wantErrorMessage: "Soundtest plate material is required.",
		},
		"ErrUserRequired": {
			user:             nil,
			url:              "/soundtests/sonnet.mp4",
			keyboard:         nil,
			keyswitch:        nil,
			keycapMaterial:   nil,
			plateMaterial:    nil,
			wantErrorCode:    clacksy.EUNAUTHORIZED,
			wantErrorMessage: "You must be logged in to create a soundtest.",
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			db := MustOpenDB(t)
			defer MustCloseDB(t, db)

			ctx := context.Background()

			ctx0 := context.Background()
			if tc.user != nil {
				_, ctx0 = MustCreateUser(t, ctx, db, tc.user, "mypassword")
			}

			var keyboard = &clacksy.Keyboard{}
			if tc.keyboard != nil {
				keyboard = MustCreateKeyboard(t, ctx, db, tc.keyboard)
			}
			var keyswitch = &clacksy.Keyswitch{}
			if tc.keyswitch != nil {
				keyswitch = MustCreateKeyswitch(t, ctx, db, tc.keyswitch)
			}
			var keycapMaterial = &clacksy.KeycapMaterial{}
			if tc.keycapMaterial != nil {
				keycapMaterial = MustCreateKeycapMaterial(t, ctx, db, tc.keycapMaterial)
			}
			var plateMaterial = &clacksy.PlateMaterial{}
			if tc.plateMaterial != nil {
				plateMaterial = MustCreatePlateMaterial(t, ctx, db, tc.plateMaterial)
			}

			s := sqlite.NewSoundtestService(db)
			soundtest := &clacksy.Soundtest{
				URL:              tc.url,
				KeyboardID:       keyboard.KeyboardID,
				KeyswitchID:      keyswitch.KeyswitchID,
				KeycapMaterialID: keycapMaterial.KeycapMaterialID,
				PlateMaterialID:  plateMaterial.PlateMaterialID,
			}

			err := s.CreateSoundtest(ctx0, soundtest)
			if tc.wantErrorCode != "" {
				if err == nil {
					t.Fatal("expected error")
				} else if clacksy.ErrorCode(err) != tc.wantErrorCode || clacksy.ErrorMessage(err) != tc.wantErrorMessage {
					t.Fatal(err)
				}
				return
			}

			if err != nil {
				t.Fatal(err)
			} else if got, want := soundtest.SoundtestID, 1; got != want {
				t.Fatalf("SoundtestID=%v, want %v", got, want)
			} else if got, want := soundtest.UserID, 1; got != want {
				t.Fatalf("UserID=%v, want %v", got, want)
			} else if soundtest.CreatedAt.IsZero() {
				t.Fatal("expected created at")
			} else if soundtest.UpdatedAt.IsZero() {
				t.Fatal("expected updated at")
			} else if soundtest.User == nil {
				t.Fatal("expected user")
			}

			if other, err := s.FindSoundtestByID(ctx0, 1); err != nil {
				t.Fatal(err)
			} else if !cmp.Equal(soundtest, other) {
				t.Fatalf("mismatch: %#v != %#v", soundtest, other)
			}
		})
	}
}

func TestSoundtestService_UpsertVote(t *testing.T) {
	t.Parallel()
	t.Run("OK", func(t *testing.T) {
		t.Parallel()
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		ctx := context.Background()
		_, ctx0 := MustCreateUser(t, ctx, db, &clacksy.User{Name: "john", Email: "john@gmail.com"}, "mypassword")

		keyboard := MustCreateKeyboard(t, ctx, db, &clacksy.Keyboard{Name: "mode sonnet"})
		keyswitch := MustCreateKeyswitch(t, ctx, db, &clacksy.Keyswitch{Name: "boba lt", KeyswitchType: &clacksy.KeyswitchType{Name: "linear"}})
		keycapMaterial := MustCreateKeycapMaterial(t, ctx, db, &clacksy.KeycapMaterial{Name: "abs"})
		plateMaterial := MustCreatePlateMaterial(t, ctx, db, &clacksy.PlateMaterial{Name: "pom"})

		soundtest := &clacksy.Soundtest{
			URL:              "/soundtests/sonnet.mp4",
			KeyboardID:       keyboard.KeyboardID,
			KeyswitchID:      keyswitch.KeyswitchID,
			KeycapMaterialID: keycapMaterial.KeycapMaterialID,
			PlateMaterialID:  plateMaterial.PlateMaterialID,
		}
		MustCreateSoundtest(t, ctx0, db, soundtest)

		vote := &clacksy.SoundtestVote{
			SoundtestID: soundtest.SoundtestID,
			VoteType:    1,
		}

		s := sqlite.NewSoundtestService(db)

		err := s.UpsertVote(ctx0, vote)
		if err != nil {
			t.Fatal(err)
		} else if got, want := vote.SoundtestID, 1; got != want {
			t.Fatalf("SoundtestID=%v, want %v", got, want)
		} else if got, want := vote.UserID, 1; got != want {
			t.Fatalf("UserID=%v, want %v", got, want)
		} else if vote.VoteType != 1 {
			t.Fatalf("VoteType=%v, want %v", vote.VoteType, 1)
		} else if soundtest.UpdatedAt.IsZero() {
			t.Fatal("expected updated at")
		}
	})

	t.Run("Update", func(t *testing.T) {
		t.Parallel()
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		ctx := context.Background()
		_, ctx0 := MustCreateUser(t, ctx, db, &clacksy.User{Name: "john", Email: "john@gmail.com"}, "mypassword")

		keyboard := MustCreateKeyboard(t, ctx, db, &clacksy.Keyboard{Name: "mode sonnet"})
		keyswitch := MustCreateKeyswitch(t, ctx, db, &clacksy.Keyswitch{Name: "boba lt", KeyswitchType: &clacksy.KeyswitchType{Name: "linear"}})
		keycapMaterial := MustCreateKeycapMaterial(t, ctx, db, &clacksy.KeycapMaterial{Name: "abs"})
		plateMaterial := MustCreatePlateMaterial(t, ctx, db, &clacksy.PlateMaterial{Name: "pom"})

		soundtest := &clacksy.Soundtest{
			URL:              "/soundtests/sonnet.mp4",
			KeyboardID:       keyboard.KeyboardID,
			KeyswitchID:      keyswitch.KeyswitchID,
			KeycapMaterialID: keycapMaterial.KeycapMaterialID,
			PlateMaterialID:  plateMaterial.PlateMaterialID,
		}
		MustCreateSoundtest(t, ctx0, db, soundtest)

		vote := &clacksy.SoundtestVote{
			SoundtestID: soundtest.SoundtestID,
			VoteType:    1,
		}

		s := sqlite.NewSoundtestService(db)

		err := s.UpsertVote(ctx0, vote)
		if err != nil {
			t.Fatal(err)
		}

		vote = &clacksy.SoundtestVote{
			SoundtestID: soundtest.SoundtestID,
			VoteType:    -1,
		}

		err = s.UpsertVote(ctx0, vote)
		if err != nil {
			t.Fatal(err)
		} else if got, want := vote.SoundtestID, 1; got != want {
			t.Fatalf("SoundtestID=%v, want %v", got, want)
		} else if got, want := vote.UserID, 1; got != want {
			t.Fatalf("UserID=%v, want %v", got, want)
		} else if vote.VoteType != -1 {
			t.Fatalf("VoteType=%v, want %v", vote.VoteType, -1)
		} else if soundtest.UpdatedAt.IsZero() {
			t.Fatal("expected updated at")
		}
	})

	t.Run("ErrSoundtestRequired", func(t *testing.T) {
		t.Parallel()
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		ctx := context.Background()
		_, ctx0 := MustCreateUser(t, ctx, db, &clacksy.User{Name: "john", Email: "john@gmail.com"}, "mypassword")

		vote := &clacksy.SoundtestVote{
			SoundtestID: 0,
			VoteType:    0,
		}

		s := sqlite.NewSoundtestService(db)

		err := s.UpsertVote(ctx0, vote)
		if err == nil {
			t.Fatal("expected error")
		} else if clacksy.ErrorCode(err) != clacksy.EINVALID || clacksy.ErrorMessage(err) != "Soundtest ID is required." {
			t.Fatal(err)
		}
	})

	t.Run("ErrVoteTypeInvalid", func(t *testing.T) {
		t.Parallel()
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		ctx := context.Background()
		_, ctx0 := MustCreateUser(t, ctx, db, &clacksy.User{Name: "john", Email: "john@gmail.com"}, "mypassword")

		keyboard := MustCreateKeyboard(t, ctx, db, &clacksy.Keyboard{Name: "mode sonnet"})
		keyswitch := MustCreateKeyswitch(t, ctx, db, &clacksy.Keyswitch{Name: "boba lt", KeyswitchType: &clacksy.KeyswitchType{Name: "linear"}})
		keycapMaterial := MustCreateKeycapMaterial(t, ctx, db, &clacksy.KeycapMaterial{Name: "abs"})
		plateMaterial := MustCreatePlateMaterial(t, ctx, db, &clacksy.PlateMaterial{Name: "pom"})

		soundtest := &clacksy.Soundtest{
			URL:              "/soundtests/sonnet.mp4",
			KeyboardID:       keyboard.KeyboardID,
			KeyswitchID:      keyswitch.KeyswitchID,
			KeycapMaterialID: keycapMaterial.KeycapMaterialID,
			PlateMaterialID:  plateMaterial.PlateMaterialID,
		}
		MustCreateSoundtest(t, ctx0, db, soundtest)

		vote := &clacksy.SoundtestVote{
			SoundtestID: soundtest.SoundtestID,
			VoteType:    -2,
		}

		s := sqlite.NewSoundtestService(db)

		err := s.UpsertVote(ctx0, vote)
		if err == nil {
			t.Fatal("expected error")
		} else if clacksy.ErrorCode(err) != clacksy.EINVALID || clacksy.ErrorMessage(err) != "Soundtest vote type must be -1, 0, or 1." {
			t.Fatal(err)
		}
	})

	t.Run("ErrUserInvalid", func(t *testing.T) {
		t.Parallel()
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		ctx := context.Background()
		_, ctx0 := MustCreateUser(t, ctx, db, &clacksy.User{Name: "john", Email: "john@gmail.com"}, "mypassword")

		keyboard := MustCreateKeyboard(t, ctx, db, &clacksy.Keyboard{Name: "mode sonnet"})
		keyswitch := MustCreateKeyswitch(t, ctx, db, &clacksy.Keyswitch{Name: "boba lt", KeyswitchType: &clacksy.KeyswitchType{Name: "linear"}})
		keycapMaterial := MustCreateKeycapMaterial(t, ctx, db, &clacksy.KeycapMaterial{Name: "abs"})
		plateMaterial := MustCreatePlateMaterial(t, ctx, db, &clacksy.PlateMaterial{Name: "pom"})

		soundtest := &clacksy.Soundtest{
			URL:              "/soundtests/sonnet.mp4",
			KeyboardID:       keyboard.KeyboardID,
			KeyswitchID:      keyswitch.KeyswitchID,
			KeycapMaterialID: keycapMaterial.KeycapMaterialID,
			PlateMaterialID:  plateMaterial.PlateMaterialID,
		}
		MustCreateSoundtest(t, ctx0, db, soundtest)

		vote := &clacksy.SoundtestVote{
			SoundtestID: soundtest.SoundtestID,
			VoteType:    0,
		}

		s := sqlite.NewSoundtestService(db)

		err := s.UpsertVote(ctx, vote)
		if err == nil {
			t.Fatal("expected error")
		} else if clacksy.ErrorCode(err) != clacksy.EUNAUTHORIZED || clacksy.ErrorMessage(err) != "You must be logged in to vote." {
			t.Fatal(err)
		}
	})
}

func TestSoundtestService_FindSoundtests(t *testing.T) {
	t.Parallel()
	t.Run("OK", func(t *testing.T) {
		t.Parallel()
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		ctx := context.Background()
		_, ctx0 := MustCreateUser(t, ctx, db, &clacksy.User{Name: "john", Email: "john@gmail.com"}, "mypassword")
		_, ctx1 := MustCreateUser(t, ctx, db, &clacksy.User{Name: "jane", Email: "jane@gmail.com"}, "mypassword")

		keyboard := MustCreateKeyboard(t, ctx, db, &clacksy.Keyboard{Name: "mode sonnet"})
		keyswitch := MustCreateKeyswitch(t, ctx, db, &clacksy.Keyswitch{Name: "boba lt", KeyswitchType: &clacksy.KeyswitchType{Name: "linear"}})
		keycapMaterial := MustCreateKeycapMaterial(t, ctx, db, &clacksy.KeycapMaterial{Name: "abs"})
		plateMaterial := MustCreatePlateMaterial(t, ctx, db, &clacksy.PlateMaterial{Name: "pom"})
		soundtest := &clacksy.Soundtest{
			URL:              "/soundtests/sonnet.mp4",
			KeyboardID:       keyboard.KeyboardID,
			KeyswitchID:      keyswitch.KeyswitchID,
			KeycapMaterialID: keycapMaterial.KeycapMaterialID,
			PlateMaterialID:  plateMaterial.PlateMaterialID,
		}

		MustCreateSoundtest(t, ctx0, db, soundtest)
		MustCreateSoundtest(t, ctx0, db, soundtest)
		MustCreateSoundtest(t, ctx1, db, soundtest)

		s := sqlite.NewSoundtestService(db)

		st, n, err := s.FindSoundtests(ctx0, clacksy.SoundtestFilter{})
		if err != nil {
			t.Fatal(err)
		} else if got, want := len(st), 3; got != want {
			t.Fatalf("len=%v, want %v", got, want)
		} else if got, want := n, 3; got != want {
			t.Fatalf("n=%v, want %v", got, want)
		}
	})

	t.Run("ByUserID", func(t *testing.T) {
		t.Parallel()
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		ctx := context.Background()
		u1, ctx0 := MustCreateUser(t, ctx, db, &clacksy.User{Name: "john", Email: "john@gmail.com"}, "mypassword")
		_, ctx1 := MustCreateUser(t, ctx, db, &clacksy.User{Name: "jane", Email: "jane@gmail.com"}, "mypassword")

		keyboard := MustCreateKeyboard(t, ctx, db, &clacksy.Keyboard{Name: "mode sonnet"})
		keyswitch := MustCreateKeyswitch(t, ctx, db, &clacksy.Keyswitch{Name: "boba lt", KeyswitchType: &clacksy.KeyswitchType{Name: "linear"}})
		keycapMaterial := MustCreateKeycapMaterial(t, ctx, db, &clacksy.KeycapMaterial{Name: "abs"})
		plateMaterial := MustCreatePlateMaterial(t, ctx, db, &clacksy.PlateMaterial{Name: "pom"})
		st1 := &clacksy.Soundtest{
			URL:              "/soundtests/st1.mp4",
			KeyboardID:       keyboard.KeyboardID,
			KeyswitchID:      keyswitch.KeyswitchID,
			KeycapMaterialID: keycapMaterial.KeycapMaterialID,
			PlateMaterialID:  plateMaterial.PlateMaterialID,
		}
		st2 := &clacksy.Soundtest{
			URL:              "/soundtests/st2.mp4",
			KeyboardID:       keyboard.KeyboardID,
			KeyswitchID:      keyswitch.KeyswitchID,
			KeycapMaterialID: keycapMaterial.KeycapMaterialID,
			PlateMaterialID:  plateMaterial.PlateMaterialID,
		}
		st3 := &clacksy.Soundtest{
			URL:              "/soundtests/st3.mp4",
			KeyboardID:       keyboard.KeyboardID,
			KeyswitchID:      keyswitch.KeyswitchID,
			KeycapMaterialID: keycapMaterial.KeycapMaterialID,
			PlateMaterialID:  plateMaterial.PlateMaterialID,
		}

		MustCreateSoundtest(t, ctx0, db, st1)
		MustCreateSoundtest(t, ctx0, db, st2)
		MustCreateSoundtest(t, ctx1, db, st3)

		s := sqlite.NewSoundtestService(db)

		st, n, err := s.FindSoundtests(ctx0, clacksy.SoundtestFilter{UserID: &u1.UserID})
		if err != nil {
			t.Fatal(err)
		} else if got, want := len(st), 2; got != want {
			t.Fatalf("len=%v, want %v", got, want)
		} else if got, want := n, 2; got != want {
			t.Fatalf("n=%v, want %v", got, want)
		} else if got, want := st[0].URL, "/soundtests/st1.mp4"; got != want {
			t.Fatalf("[0]=%v, want %v", got, want)
		} else if got, want := st[1].URL, "/soundtests/st2.mp4"; got != want {
			t.Fatalf("[1]=%v, want %v", got, want)
		}
	})

	t.Run("BySoundtestID", func(t *testing.T) {
		t.Parallel()
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		ctx := context.Background()
		_, ctx0 := MustCreateUser(t, ctx, db, &clacksy.User{Name: "john", Email: "john@gmail.com"}, "mypassword")

		keyboard := MustCreateKeyboard(t, ctx, db, &clacksy.Keyboard{Name: "mode sonnet"})
		keyswitch := MustCreateKeyswitch(t, ctx, db, &clacksy.Keyswitch{Name: "boba lt", KeyswitchType: &clacksy.KeyswitchType{Name: "linear"}})
		keycapMaterial := MustCreateKeycapMaterial(t, ctx, db, &clacksy.KeycapMaterial{Name: "abs"})
		plateMaterial := MustCreatePlateMaterial(t, ctx, db, &clacksy.PlateMaterial{Name: "pom"})
		st1 := &clacksy.Soundtest{
			URL:              "/soundtests/st1.mp4",
			KeyboardID:       keyboard.KeyboardID,
			KeyswitchID:      keyswitch.KeyswitchID,
			KeycapMaterialID: keycapMaterial.KeycapMaterialID,
			PlateMaterialID:  plateMaterial.PlateMaterialID,
		}
		st2 := &clacksy.Soundtest{
			URL:              "/soundtests/st2.mp4",
			KeyboardID:       keyboard.KeyboardID,
			KeyswitchID:      keyswitch.KeyswitchID,
			KeycapMaterialID: keycapMaterial.KeycapMaterialID,
			PlateMaterialID:  plateMaterial.PlateMaterialID,
		}

		s1 := MustCreateSoundtest(t, ctx0, db, st1)
		MustCreateSoundtest(t, ctx0, db, st2)

		s := sqlite.NewSoundtestService(db)

		st, n, err := s.FindSoundtests(ctx0, clacksy.SoundtestFilter{SoundtestID: &s1.SoundtestID})
		if err != nil {
			t.Fatal(err)
		} else if got, want := len(st), 1; got != want {
			t.Fatalf("len=%v, want %v", got, want)
		} else if got, want := n, 1; got != want {
			t.Fatalf("n=%v, want %v", got, want)
		} else if got, want := st[0].URL, "/soundtests/st1.mp4"; got != want {
			t.Fatalf("[0]=%v, want %v", got, want)
		}
	})

	t.Run("LimitOffset", func(t *testing.T) {
		t.Parallel()
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		ctx := context.Background()
		_, ctx0 := MustCreateUser(t, ctx, db, &clacksy.User{Name: "john", Email: "john@gmail.com"}, "mypassword")

		keyboard := MustCreateKeyboard(t, ctx, db, &clacksy.Keyboard{Name: "mode sonnet"})
		keyswitch := MustCreateKeyswitch(t, ctx, db, &clacksy.Keyswitch{Name: "boba lt", KeyswitchType: &clacksy.KeyswitchType{Name: "linear"}})
		keycapMaterial := MustCreateKeycapMaterial(t, ctx, db, &clacksy.KeycapMaterial{Name: "abs"})
		plateMaterial := MustCreatePlateMaterial(t, ctx, db, &clacksy.PlateMaterial{Name: "pom"})
		st1 := &clacksy.Soundtest{
			URL:              "/soundtests/st1.mp4",
			KeyboardID:       keyboard.KeyboardID,
			KeyswitchID:      keyswitch.KeyswitchID,
			KeycapMaterialID: keycapMaterial.KeycapMaterialID,
			PlateMaterialID:  plateMaterial.PlateMaterialID,
		}
		st2 := &clacksy.Soundtest{
			URL:              "/soundtests/st2.mp4",
			KeyboardID:       keyboard.KeyboardID,
			KeyswitchID:      keyswitch.KeyswitchID,
			KeycapMaterialID: keycapMaterial.KeycapMaterialID,
			PlateMaterialID:  plateMaterial.PlateMaterialID,
		}

		MustCreateSoundtest(t, ctx0, db, st1)
		MustCreateSoundtest(t, ctx0, db, st2)

		s := sqlite.NewSoundtestService(db)

		st, n, err := s.FindSoundtests(ctx0, clacksy.SoundtestFilter{Limit: 1, Offset: 1})
		if err != nil {
			t.Fatal(err)
		} else if got, want := len(st), 1; got != want {
			t.Fatalf("len=%v, want %v", got, want)
		} else if got, want := n, 2; got != want {
			t.Fatalf("n=%v, want %v", got, want)
		} else if got, want := st[0].URL, "/soundtests/st2.mp4"; got != want {
			t.Fatalf("[0]=%v, want %v", got, want)
		}
	})
}

func TestSoundtestService_FindKeyboards(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		ctx := context.Background()
		k1 := MustCreateKeyboard(t, ctx, db, &clacksy.Keyboard{Name: "mode sonnet"})
		k2 := MustCreateKeyboard(t, ctx, db, &clacksy.Keyboard{Name: "unikorn"})

		s := sqlite.NewSoundtestService(db)
		keebs, err := s.FindKeyboards(ctx, clacksy.KeyboardFilter{})
		if err != nil {
			t.Fatal(err)
		} else if got, want := len(keebs), 2; got != want {
			t.Fatalf("len=%v, want %v", got, want)
		} else if !cmp.Equal(keebs[0], k1) {
			t.Fatalf("mismatch: %#v != %#v", keebs[0], k1)
		} else if !cmp.Equal(keebs[1], k2) {
			t.Fatalf("mismatch: %#v != %#v", keebs[1], k2)
		}
	})

	t.Run("Filter", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		ctx := context.Background()
		k1 := MustCreateKeyboard(t, ctx, db, &clacksy.Keyboard{Name: "mode sonnet"})
		for i := 0; i < 10; i++ {
			MustCreateKeyboard(t, ctx, db, &clacksy.Keyboard{Name: strconv.Itoa(i)})
		}

		s := sqlite.NewSoundtestService(db)
		keebs, err := s.FindKeyboards(ctx, clacksy.KeyboardFilter{KeyboardID: &k1.KeyboardID})
		if err != nil {
			t.Fatal(err)
		} else if got, want := len(keebs), 4; got != want {
			t.Fatalf("len=%v, want %v", got, want)
		}

		contains := false
		for _, keeb := range keebs {
			if keeb.KeyboardID == k1.KeyboardID {
				contains = true
			}
		}
		if !contains {
			t.Fatalf("correct keyboard: %v is missing", k1.KeyboardID)
		}
	})
}

func TestSoundtestService_FindKeyswitches(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		ctx := context.Background()
		k1 := MustCreateKeyswitch(t, ctx, db, &clacksy.Keyswitch{Name: "boba lt", KeyswitchType: &clacksy.KeyswitchType{Name: "linear"}})
		k2 := MustCreateKeyswitch(t, ctx, db, &clacksy.Keyswitch{Name: "gateron yellow", KeyswitchType: &clacksy.KeyswitchType{Name: "tactile"}})

		s := sqlite.NewSoundtestService(db)
		switches, err := s.FindKeyswitches(ctx, clacksy.KeyswitchFilter{})
		if err != nil {
			t.Fatal(err)
		} else if got, want := len(switches), 2; got != want {
			t.Fatalf("len=%v, want %v", got, want)
		} else if !cmp.Equal(switches[0], k1) {
			t.Fatalf("mismatch: %#v != %#v", switches[0], k1)
		} else if !cmp.Equal(switches[1], k2) {
			t.Fatalf("mismatch: %#v != %#v", switches[1], k2)
		}
	})

	t.Run("Filter", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		ctx := context.Background()
		k1 := MustCreateKeyswitch(t, ctx, db, &clacksy.Keyswitch{Name: "boba lt", KeyswitchType: &clacksy.KeyswitchType{Name: "linear"}})
		for i := 0; i < 10; i++ {
			MustCreateKeyswitch(t, ctx, db, &clacksy.Keyswitch{Name: strconv.Itoa(i), KeyswitchType: &clacksy.KeyswitchType{Name: "linear"}})
		}

		s := sqlite.NewSoundtestService(db)
		switches, err := s.FindKeyswitches(ctx, clacksy.KeyswitchFilter{KeyswitchID: &k1.KeyswitchID})
		if err != nil {
			t.Fatal(err)
		} else if got, want := len(switches), 4; got != want {
			t.Fatalf("len=%v, want %v", got, want)
		}

		contains := false
		for _, sw := range switches {
			if sw.KeyswitchID == k1.KeyswitchID {
				contains = true
			}
		}
		if !contains {
			t.Fatalf("correct keyswitch: %v is missing", k1.KeyswitchID)
		}
	})
}

func TestSoundtestService_FindPlateMaterials(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		ctx := context.Background()
		p1 := MustCreatePlateMaterial(t, ctx, db, &clacksy.PlateMaterial{Name: "alu"})
		p2 := MustCreatePlateMaterial(t, ctx, db, &clacksy.PlateMaterial{Name: "pom"})

		s := sqlite.NewSoundtestService(db)
		plateMaterials, err := s.FindPlateMaterials(ctx, clacksy.PlateMaterialFilter{})
		if err != nil {
			t.Fatal(err)
		} else if got, want := len(plateMaterials), 2; got != want {
			t.Fatalf("len=%v, want %v", got, want)
		} else if !cmp.Equal(plateMaterials[0], p1) {
			t.Fatalf("mismatch: %#v != %#v", plateMaterials[0], p1)
		} else if !cmp.Equal(plateMaterials[1], p2) {
			t.Fatalf("mismatch: %#v != %#v", plateMaterials[1], p2)
		}
	})

	t.Run("Filter", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		ctx := context.Background()
		p1 := MustCreatePlateMaterial(t, ctx, db, &clacksy.PlateMaterial{Name: "alu"})
		for i := 0; i < 10; i++ {
			MustCreatePlateMaterial(t, ctx, db, &clacksy.PlateMaterial{Name: strconv.Itoa(i)})
		}

		s := sqlite.NewSoundtestService(db)
		plateMaterials, err := s.FindPlateMaterials(ctx, clacksy.PlateMaterialFilter{PlateMaterialID: &p1.PlateMaterialID})
		if err != nil {
			t.Fatal(err)
		} else if got, want := len(plateMaterials), 4; got != want {
			t.Fatalf("len=%v, want %v", got, want)
		}

		contains := false
		for _, pm := range plateMaterials {
			if pm.PlateMaterialID == p1.PlateMaterialID {
				contains = true
			}
		}
		if !contains {
			t.Fatalf("correct plate material: %v is missing", p1.PlateMaterialID)
		}
	})
}

func TestSoundtestService_FindKeycapMaterials(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		ctx := context.Background()
		k1 := MustCreateKeycapMaterial(t, ctx, db, &clacksy.KeycapMaterial{Name: "abs"})
		k2 := MustCreateKeycapMaterial(t, ctx, db, &clacksy.KeycapMaterial{Name: "pbt"})

		s := sqlite.NewSoundtestService(db)
		keycapMaterials, err := s.FindKeycapMaterials(ctx, clacksy.KeycapMaterialFilter{})
		if err != nil {
			t.Fatal(err)
		} else if got, want := len(keycapMaterials), 2; got != want {
			t.Fatalf("len=%v, want %v", got, want)
		} else if !cmp.Equal(keycapMaterials[0], k1) {
			t.Fatalf("mismatch: %#v != %#v", keycapMaterials[0], k1)
		} else if !cmp.Equal(keycapMaterials[1], k2) {
			t.Fatalf("mismatch: %#v != %#v", keycapMaterials[1], k2)
		}
	})

	t.Run("Filter", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		ctx := context.Background()
		k1 := MustCreateKeycapMaterial(t, ctx, db, &clacksy.KeycapMaterial{Name: "abs"})
		for i := 0; i < 10; i++ {
			MustCreateKeycapMaterial(t, ctx, db, &clacksy.KeycapMaterial{Name: strconv.Itoa(i)})
		}

		s := sqlite.NewSoundtestService(db)
		keycapMaterials, err := s.FindKeycapMaterials(ctx, clacksy.KeycapMaterialFilter{KeycapMaterialID: &k1.KeycapMaterialID})
		if err != nil {
			t.Fatal(err)
		} else if got, want := len(keycapMaterials), 4; got != want {
			t.Fatalf("len=%v, want %v", got, want)
		}

		contains := false
		for _, km := range keycapMaterials {
			if km.KeycapMaterialID == k1.KeycapMaterialID {
				contains = true
			}
		}
		if !contains {
			t.Fatalf("correct keycap material: %v is missing", k1.KeycapMaterialID)
		}
	})
}

func MustFindSoundtestByID(tb testing.TB, ctx context.Context, db *sqlite.DB, soundtestID int) *clacksy.Soundtest {
	tb.Helper()
	soundtest, err := sqlite.NewSoundtestService(db).FindSoundtestByID(ctx, soundtestID)
	if err != nil {
		tb.Fatal(err)
	}

	return soundtest
}

func MustCreateSoundtest(tb testing.TB, ctx context.Context, db *sqlite.DB, soundtest *clacksy.Soundtest) *clacksy.Soundtest {
	tb.Helper()
	if err := sqlite.NewSoundtestService(db).CreateSoundtest(ctx, soundtest); err != nil {
		tb.Fatal(err)
	}

	return soundtest
}

func MustCreateKeyboard(tb testing.TB, ctx context.Context, db *sqlite.DB, keyboard *clacksy.Keyboard) *clacksy.Keyboard {
	tb.Helper()

	err := sqlite.NewSoundtestService(db).CreateKeyboard(ctx, keyboard)
	if err != nil {
		tb.Fatal(err)
	}

	return keyboard
}

func MustCreateKeyswitch(tb testing.TB, ctx context.Context, db *sqlite.DB, keyswitch *clacksy.Keyswitch) *clacksy.Keyswitch {
	tb.Helper()

	err := sqlite.NewSoundtestService(db).CreateKeyswitch(ctx, keyswitch)
	if err != nil {
		tb.Fatal(err)
	}

	return keyswitch
}

func MustCreateKeycapMaterial(tb testing.TB, ctx context.Context, db *sqlite.DB, keycapMaterial *clacksy.KeycapMaterial) *clacksy.KeycapMaterial {
	tb.Helper()

	err := sqlite.NewSoundtestService(db).CreateKeycapMaterial(ctx, keycapMaterial)
	if err != nil {
		tb.Fatal(err)
	}

	return keycapMaterial
}

func MustCreatePlateMaterial(tb testing.TB, ctx context.Context, db *sqlite.DB, plateMaterial *clacksy.PlateMaterial) *clacksy.PlateMaterial {
	tb.Helper()

	err := sqlite.NewSoundtestService(db).CreatePlateMaterial(ctx, plateMaterial)
	if err != nil {
		tb.Fatal(err)
	}

	return plateMaterial
}
