package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/0xhjohnson/clacksy/models"
)

func TestFilenameWithoutExt(t *testing.T) {
	tests := map[string]struct {
		filename string
		want     string
	}{
		".mp4 filename": {
			filename: "soundtest-3483310207.mp4",
			want:     "soundtest-3483310207",
		},
		".mov filename": {
			filename: "somefile.mov",
			want:     "somefile",
		},
		"includes path": {
			filename: "/var/folders/x6/soundtest.aac",
			want:     "/var/folders/x6/soundtest",
		},
		"random chars": {
			filename: "1xq_-$.mp3",
			want:     "1xq_-$",
		},
		"unknown extension": {
			filename: "songs.xyz",
			want:     "songs",
		},
		"simple": {
			filename: "file",
			want:     "file",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := filenameWithoutExt(tc.filename)
			if tc.want != got {
				t.Errorf("want: %v, got: %v", tc.want, got)
			}
		})
	}
}

func TestHasPlayed(t *testing.T) {
	app := &application{}

	tests := map[string]struct {
		r       *http.Request
		context context.Context
		want    bool
	}{
		"has played": {
			r:       &http.Request{},
			context: context.WithValue(context.Background(), userPlayContextKey, &models.SoundTestPlay{}),
			want:    true,
		},
		"context key is nil": {
			r:       &http.Request{},
			context: context.WithValue(context.Background(), userPlayContextKey, nil),
			want:    false,
		},
		"fresh context": {
			r:       &http.Request{},
			context: context.Background(),
			want:    false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := app.hasPlayed(tc.r.WithContext(tc.context))
			if tc.want != got {
				t.Errorf("want: %t, got: %t", tc.want, got)
			}
		})
	}
}

func TestIsAuthenticated(t *testing.T) {
	app := &application{}

	tests := map[string]struct {
		r       *http.Request
		context context.Context
		want    bool
	}{
		"is authenticated": {
			r:       &http.Request{},
			context: context.WithValue(context.Background(), isAuthenticatedContextKey, true),
			want:    true,
		},
		"context key is nil": {
			r:       &http.Request{},
			context: context.WithValue(context.Background(), isAuthenticatedContextKey, nil),
			want:    false,
		},
		"context key is false": {
			r:       &http.Request{},
			context: context.WithValue(context.Background(), isAuthenticatedContextKey, false),
			want:    false,
		},
		"fresh context": {
			r:       &http.Request{},
			context: context.Background(),
			want:    false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := app.isAuthenticated(tc.r.WithContext(tc.context))
			if tc.want != got {
				t.Errorf("want: %t, got: %t", tc.want, got)
			}
		})
	}
}

func TestClientError(t *testing.T) {
	app := &application{}

	tests := map[string]struct {
		statusCode int
		want       string
	}{
		"bad request": {
			statusCode: http.StatusBadRequest,
			want:       http.StatusText(http.StatusBadRequest),
		},
		"unauthorized": {
			statusCode: http.StatusUnauthorized,
			want:       http.StatusText(http.StatusUnauthorized),
		},
		"teapot": {
			statusCode: http.StatusTeapot,
			want:       http.StatusText(http.StatusTeapot),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			app.clientError(w, tc.statusCode)

			if w.Code != tc.statusCode {
				t.Errorf("want statusCode %d, got %d", tc.statusCode, w.Code)
			}
			if strings.TrimSpace(w.Body.String()) != tc.want {
				t.Errorf("want body of %s, got %s", tc.want, w.Body.String())
			}
		})
	}
}

func TestServerError(t *testing.T) {
	var buf bytes.Buffer

	app := &application{errorLog: log.New(&buf, "", 0)}

	tests := map[string]struct {
		err  error
		want string
	}{
		"database error": {
			err:  fmt.Errorf("error connecting to db"),
			want: "error connecting to db",
		},
		"auth error": {
			err:  fmt.Errorf("error authenticating user"),
			want: "error authenticating user",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			app.serverError(w, tc.err)

			if w.Code != http.StatusInternalServerError {
				t.Errorf("want internal server error statusCode %d, got %d", http.StatusInternalServerError, w.Code)
			}
			if !strings.Contains(buf.String(), tc.want) {
				t.Errorf("want log message of %s, got %s", tc.want, buf.String())
			}
		})
	}
}
