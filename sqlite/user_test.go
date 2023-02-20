package sqlite_test

import (
	"context"
	"testing"

	"github.com/0xhjohnson/clacksy"
	"github.com/0xhjohnson/clacksy/sqlite"
)

func TestUserService_CreateUser(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewUserService(db)

		u := &clacksy.User{
			Email: "hi@0xhjohnson.com",
		}
		password := "someadvancedpassword"

		if err := s.CreateUser(context.Background(), u, password); err != nil {
			t.Fatal(err)
		} else if got, want := u.UserID, 1; got != want {
			t.Fatalf("UserID=%v, want %v", got, want)
		} else if u.CreatedAt.IsZero() {
			t.Fatal("expected created at")
		} else if u.UpdatedAt.IsZero() {
			t.Fatal("expected created at")
		}

		u2 := &clacksy.User{
			Email: "somebody@gmail.com",
		}

		if err := s.CreateUser(context.Background(), u2, "testing123"); err != nil {
			t.Fatal(err)
		} else if got, want := u2.UserID, 2; got != want {
			t.Fatalf("UserID=%v, want %v", got, want)
		}
	})

	t.Run("ErrEmailRequired", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewUserService(db)

		err := s.CreateUser(context.Background(), &clacksy.User{}, "mypassword")
		if err == nil {
			t.Fatal("expected error")
		} else if clacksy.ErrorCode(err) != clacksy.EINVALID || clacksy.ErrorMessage(err) != "User email is required." {
			t.Fatalf("unexpected error: %#v", err)
		}
	})

	t.Run("ErrPasswordRequired", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewUserService(db)

		u := &clacksy.User{
			Email: "somebody@gmail.com",
		}

		err := s.CreateUser(context.Background(), u, "")
		if err == nil {
			t.Fatal("expected error")
		} else if clacksy.ErrorCode(err) != clacksy.EINVALID || clacksy.ErrorMessage(err) != "User password is required." {
			t.Fatalf("unexpected error: %#v", err)
		}
	})

	t.Run("ErrPasswordMinLength", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewUserService(db)

		u := &clacksy.User{
			Email: "somebody@gmail.com",
		}

		err := s.CreateUser(context.Background(), u, "short")
		if err == nil {
			t.Fatal("expected error")
		} else if clacksy.ErrorCode(err) != clacksy.EINVALID || clacksy.ErrorMessage(err) != "User password must be at least 8 characters." {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
}
