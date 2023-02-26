package sqlite_test

import (
	"context"
	"testing"

	"github.com/0xhjohnson/clacksy"
	"github.com/0xhjohnson/clacksy/sqlite"
	"github.com/google/go-cmp/cmp"
)

func TestUserService_CreateUser(t *testing.T) {
	t.Parallel()
	t.Run("OK", func(t *testing.T) {
		t.Parallel()
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewUserService(db)

		u := &clacksy.User{
			Email: "hi@0xhjohnson.com",
		}
		password := "someadvancedpassword"

		err := s.CreateUser(context.Background(), u, password)
		if err != nil {
			t.Fatal(err)
		} else if got, want := u.UserID, 1; got != want {
			t.Fatalf("UserID=%v, want %v", got, want)
		} else if u.CreatedAt.IsZero() {
			t.Fatal("expected created at")
		} else if u.UpdatedAt.IsZero() {
			t.Fatal("expected updated at")
		}

		u2 := &clacksy.User{
			Email: "somebody@gmail.com",
		}

		err = s.CreateUser(context.Background(), u2, "testing123")
		if err != nil {
			t.Fatal(err)
		} else if got, want := u2.UserID, 2; got != want {
			t.Fatalf("UserID=%v, want %v", got, want)
		}

		other, err := s.FindUserByID(context.Background(), 1)
		if err != nil {
			t.Fatal(err)
		} else if !cmp.Equal(u, other) {
			t.Fatalf("mismatch: %#v != %#v", u, other)
		}
	})

	t.Run("ErrEmailRequired", func(t *testing.T) {
		t.Parallel()
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
		t.Parallel()
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
		t.Parallel()
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

func TestUserService_UpdateUser(t *testing.T) {
	t.Parallel()
	t.Run("OK", func(t *testing.T) {
		t.Parallel()
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewUserService(db)

		user0, ctx0 := MustCreateUser(t, context.Background(), db, &clacksy.User{
			Email: "hi@0xhjohnson.com",
			Name:  "hunter",
		}, "agoodpassword")

		newName, newEmail := "hunt", "hunt@gmail.com"
		uu, err := s.UpdateUser(ctx0, user0.UserID, clacksy.UserUpdate{
			Email: &newEmail,
			Name:  &newName,
		})
		if err != nil {
			t.Fatal(err)
		} else if got, want := uu.Name, "hunt"; got != want {
			t.Fatalf("Name=%v, want %v", got, want)
		} else if got, want := uu.Email, "hunt@gmail.com"; got != want {
			t.Fatalf("Email=%v, want %v", got, want)
		}

		other, err := s.FindUserByID(context.Background(), 1)
		if err != nil {
			t.Fatal(err)
		} else if !cmp.Equal(uu, other) {
			t.Fatalf("mismatch: %#v != %#v", uu, other)
		}
	})

	t.Run("ErrUnauthorized", func(t *testing.T) {
		t.Parallel()
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewUserService(db)

		user0, _ := MustCreateUser(t, context.Background(), db, &clacksy.User{Email: "bob@gmail.com"}, "bobpassword")
		_, ctx1 := MustCreateUser(t, context.Background(), db, &clacksy.User{Email: "rob@gmail.com"}, "robpassword")

		newName := "newname"

		_, err := s.UpdateUser(ctx1, user0.UserID, clacksy.UserUpdate{Name: &newName})
		if err == nil {
			t.Fatal("expected error")
		} else if clacksy.ErrorCode(err) != clacksy.EUNAUTHORIZED || clacksy.ErrorMessage(err) != `You are not allowed to update this user.` {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
}

func TestUserService_Authenticate(t *testing.T) {
	t.Parallel()
	t.Run("OK", func(t *testing.T) {
		t.Parallel()
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewUserService(db)

		name := "hunter"
		email := "hunter@gmail.com"
		password := "agoodpassword"
		u, ctx := MustCreateUser(t, context.Background(), db, &clacksy.User{
			Email: email,
			Name:  name,
		}, password)

		user, err := s.Authenticate(ctx, u, password)
		if err != nil {
			t.Fatal(err)
		} else if got, want := user.Email, email; got != want {
			t.Fatalf("Email=%v, want %v", got, want)
		} else if got, want := user.Name, name; got != want {
			t.Fatalf("Name=%v, want %v", got, want)
		}
	})

	t.Run("ErrNotFound", func(t *testing.T) {
		t.Parallel()
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewUserService(db)

		email := "hunter@gmail.com"
		password := "agoodpassword"
		_, ctx := MustCreateUser(t, context.Background(), db, &clacksy.User{
			Email: email,
		}, password)

		_, err := s.Authenticate(ctx, &clacksy.User{Email: "invalid@gmail.com"}, password)
		if err == nil {
			t.Fatal("expected error")
		} else if clacksy.ErrorCode(err) != clacksy.ENOTFOUND || clacksy.ErrorMessage(err) != "User not found." {
			t.Fatalf("unexpected error: %#v", err)
		}

	})

	t.Run("ErrInvalidCreds", func(t *testing.T) {
		t.Parallel()
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewUserService(db)

		name := "hunter"
		email := "hunter@gmail.com"
		password := "agoodpassword"
		u, ctx := MustCreateUser(t, context.Background(), db, &clacksy.User{
			Email: email,
			Name:  name,
		}, password)

		_, err := s.Authenticate(ctx, u, "incorrectpass")
		if err == nil {
			t.Fatal("expected error")
		} else if clacksy.ErrorCode(err) != clacksy.EINVALID || clacksy.ErrorMessage(err) != "User credentials invalid." {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
}

func TestUserService_FindUser(t *testing.T) {
	t.Parallel()
	t.Run("ErrNotFound", func(t *testing.T) {
		t.Parallel()
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewUserService(db)

		_, err := s.FindUserByID(context.Background(), 1)
		if clacksy.ErrorCode(err) != clacksy.ENOTFOUND {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
}

func MustCreateUser(tb testing.TB, ctx context.Context, db *sqlite.DB, user *clacksy.User, password string) (*clacksy.User, context.Context) {
	tb.Helper()
	if err := sqlite.NewUserService(db).CreateUser(ctx, user, password); err != nil {
		tb.Fatal(err)
	}
	return user, clacksy.NewContextWithUser(ctx, user)
}
