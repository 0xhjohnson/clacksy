package main

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gofrs/uuid"
)

func (app *application) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !app.isAuthenticated(r) {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}

		w.Header().Add("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := app.sessionManager.GetString(r.Context(), "authenticatedUserID")
		if id == "" {
			next.ServeHTTP(w, r)
			return
		}

		uID, err := uuid.FromString(id)
		if err != nil {
			app.serverError(w, err)
			return
		}

		exists, err := app.users.Exists(uID)
		if err != nil {
			app.serverError(w, err)
			return
		}

		if exists {
			ctx := context.WithValue(r.Context(), isAuthenticatedContextKey, true)
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) paginate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pageQ := r.URL.Query().Get("page")

		page := 0

		if pageQ != "" {
			var err error
			page, err = strconv.Atoi(pageQ)
			if err != nil {
				app.serverError(w, err)
				return
			}
		}

		ctx := context.WithValue(r.Context(), pageContextKey, page)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
