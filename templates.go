package main

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"text/template"

	"github.com/0xhjohnson/clacksy/ui"
	"github.com/gofrs/uuid"
)

type templateData struct {
	Form            any
	Flash           string
	StaticPath      string
	URLPath         string
	AppEnv          string
	IsAuthenticated bool
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := fs.Glob(ui.Files, "html/pages/*.tmpl")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		patterns := []string{
			"html/layout.tmpl",
			"html/partials/*.tmpl",
			page,
		}

		funcMap := template.FuncMap{
			"uuidEq": func(s string, u uuid.UUID) bool {
				return s == u.String()
			},
		}

		ts, err := template.New(name).Funcs(funcMap).ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}

func (app *application) newTemplateData(r *http.Request) *templateData {
	var staticPath string
	appEnv := os.Getenv("APP_ENV")

	if appEnv == "production" {
		staticPath = "https://cdn.clacksy.com/file/clacksy"
	} else {
		staticPath = "/static"
	}

	return &templateData{
		Flash:           app.sessionManager.PopString(r.Context(), "flash"),
		StaticPath:      staticPath,
		URLPath:         r.URL.Path,
		AppEnv:          appEnv,
		IsAuthenticated: app.isAuthenticated(r),
	}
}

func (app *application) renderTemplate(w http.ResponseWriter, statusCode int, page string, data *templateData) {
	tmpl, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.serverError(w, err)
		return
	}

	w.WriteHeader(statusCode)

	err := tmpl.ExecuteTemplate(w, "layout", data)
	if err != nil {
		app.serverError(w, err)
		return
	}
}
