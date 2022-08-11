package main

import (
	"net/http"
	"time"

	"github.com/0xhjohnson/clacksy/ui"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *application) routes() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(30 * time.Second))

	r.With(app.sessionManager.LoadAndSave).Get("/", app.home)

	r.Route("/user", func(r chi.Router) {
		r.Use(app.sessionManager.LoadAndSave)
		r.Get("/new", app.newUserForm)
		r.Post("/new", app.addNewUser)
		r.Get("/login", app.loginUserForm)
		r.Post("/login", app.loginUser)
		r.Post("/logout", app.logoutUser)
	})

	fileServer := http.FileServer(http.FS(ui.Files))
	r.Handle("/static/*", fileServer)

	return r
}
