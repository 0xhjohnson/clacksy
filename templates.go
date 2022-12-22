package main

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/0xhjohnson/clacksy/ui"
	"github.com/gofrs/uuid"
)

var staticURL string = "https://cdn.clacksy.com/file/clacksy"

type templateData struct {
	Form            any
	Flash           string
	PublicPath      string
	URLPath         string
	StaticURL       string
	AppEnv          string
	IsAuthenticated bool
	PageData        any
}

func uuidEq(s string, u uuid.UUID) bool {
	return s == u.String()
}

func humanDate(t time.Time) string {
	return t.Format("02 Jan 2006 at 1:04PM")
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
			"uuidEq":    uuidEq,
			"humanDate": humanDate,
			"hasPrefix": strings.HasPrefix,
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
	var publicPath string
	appEnv := os.Getenv("APP_ENV")

	if appEnv == "production" {
		publicPath = staticURL
	} else {
		publicPath = "/public"
	}

	return &templateData{
		Flash:           app.sessionManager.PopString(r.Context(), "flash"),
		PublicPath:      publicPath,
		StaticURL:       staticURL,
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
