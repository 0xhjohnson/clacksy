package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireAuth(t *testing.T) {
	app := application{}

	tests := map[string]struct {
		context          context.Context
		wantStatusCode   int
		wantRedirect     string
		wantCacheControl string
	}{
		"not authenticated": {
			context:          context.WithValue(context.Background(), isAuthenticatedContextKey, false),
			wantStatusCode:   http.StatusSeeOther,
			wantRedirect:     "/user/login",
			wantCacheControl: "",
		},
		"authenticated": {
			context:          context.WithValue(context.Background(), isAuthenticatedContextKey, true),
			wantStatusCode:   http.StatusOK,
			wantRedirect:     "",
			wantCacheControl: "no-store",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			rr := httptest.NewRecorder()
			req, err := http.NewRequestWithContext(tc.context, "GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			handler := app.requireAuth(next)
			handler.ServeHTTP(rr, req)

			if rr.Code != tc.wantStatusCode {
				t.Errorf("handler returned wrong status code, got: %d, want: %d", rr.Code, tc.wantStatusCode)
			}
			if rr.Result().Header.Get("Cache-Control") != tc.wantCacheControl {
				t.Errorf("handler Cache-Control header wrong, got: %s, want: %s", rr.Result().Header.Get("Cache-Control"), tc.wantCacheControl)
			}

			redirectUrl, err := rr.Result().Location()
			if err != nil {
				switch {
				case errors.Is(err, http.ErrNoLocation):
					if tc.wantRedirect == "" {
						return
					}
					t.Fatalf("handler didn't redirect, want: %s", tc.wantRedirect)
				default:
					t.Fatal(err)
				}
			}

			if redirectUrl.String() != tc.wantRedirect {
				t.Errorf("handler redirected incorrectly, got: %s, want: %s", redirectUrl, tc.wantRedirect)
			}
		})
	}
}

func TestLimitPlayOnce(t *testing.T) {
	app := application{}

	tests := map[string]struct {
		context        context.Context
		wantStatusCode int
		wantRedirect   string
	}{
		"has played": {
			context:        context.WithValue(context.Background(), userPlayContextKey, true),
			wantStatusCode: http.StatusSeeOther,
			wantRedirect:   "/play/grade",
		},
		"not played": {
			context:        context.Background(),
			wantStatusCode: http.StatusOK,
			wantRedirect:   "",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			rr := httptest.NewRecorder()
			req, err := http.NewRequestWithContext(tc.context, "GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			handler := app.limitPlayOnce(next)
			handler.ServeHTTP(rr, req)

			if rr.Code != tc.wantStatusCode {
				t.Errorf("handler returned wrong status code, got: %d, want: %d", rr.Code, tc.wantStatusCode)
			}

			redirectUrl, err := rr.Result().Location()
			if err != nil {
				switch {
				case errors.Is(err, http.ErrNoLocation):
					if tc.wantRedirect == "" {
						return
					}
					t.Fatalf("handler didn't redirect, want: %s", tc.wantRedirect)
				default:
					t.Fatal(err)
				}
			}

			if redirectUrl.String() != tc.wantRedirect {
				t.Errorf("handler redirected incorrectly, got: %s, want: %s", redirectUrl, tc.wantRedirect)
			}
		})
	}
}

func TestVerifyPlayed(t *testing.T) {
	app := application{}

	tests := map[string]struct {
		context        context.Context
		wantStatusCode int
		wantRedirect   string
	}{
		"not played": {
			context:        context.Background(),
			wantStatusCode: http.StatusSeeOther,
			wantRedirect:   "/play",
		},
		"played": {
			context:        context.WithValue(context.Background(), userPlayContextKey, true),
			wantStatusCode: http.StatusOK,
			wantRedirect:   "",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			rr := httptest.NewRecorder()
			req, err := http.NewRequestWithContext(tc.context, "GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			handler := app.verifyPlayed(next)
			handler.ServeHTTP(rr, req)

			if rr.Code != tc.wantStatusCode {
				t.Errorf("handler returned wrong status code, got: %d, want: %d", rr.Code, tc.wantStatusCode)
			}

			redirectUrl, err := rr.Result().Location()
			if err != nil {
				switch {
				case errors.Is(err, http.ErrNoLocation):
					if tc.wantRedirect == "" {
						return
					}
					t.Fatalf("handler didn't redirect, want: %s", tc.wantRedirect)
				default:
					t.Fatal(err)
				}
			}

			if redirectUrl.String() != tc.wantRedirect {
				t.Errorf("handler redirected incorrectly, got: %s, want: %s", redirectUrl, tc.wantRedirect)
			}
		})
	}
}
