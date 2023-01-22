package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"testing/fstest"
	"text/template"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/gofrs/uuid"
)

func TestUuidEq(t *testing.T) {
	tests := map[string]struct {
		s    string
		u    uuid.UUID
		want bool
	}{
		"equal": {
			s:    "05ff139b-8b9a-4341-a161-8628c3e038e7",
			u:    uuid.Must(uuid.FromString("05ff139b-8b9a-4341-a161-8628c3e038e7")),
			want: true,
		},
		"not equal": {
			s:    "05ff139b-8b9a-4341-a161-8628c3e038e7",
			u:    uuid.Must(uuid.FromString("2a0f3ad4-216d-43db-a731-fdab599e2d45")),
			want: false,
		},
		"empty": {
			s:    "",
			u:    uuid.Nil,
			want: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := uuidEq(tc.s, tc.u)
			if tc.want != got {
				t.Errorf("want: %t, got: %t", tc.want, got)
			}
		})
	}
}

func TestHumanDate(t *testing.T) {
	tests := map[string]struct {
		time time.Time
		want string
	}{
		"January 2, 2021 at 1:04pm": {
			time: time.Date(2021, time.January, 2, 13, 4, 0, 0, time.UTC),
			want: "02 Jan 2021 at 1:04PM",
		},
		"December 30, 2020 at 11:59pm": {
			time: time.Date(2020, time.December, 30, 23, 59, 0, 0, time.UTC),
			want: "30 Dec 2020 at 11:59PM",
		},
		"February 14, 2022 at 2:30AM": {
			time: time.Date(2022, time.February, 14, 2, 30, 0, 0, time.UTC),
			want: "14 Feb 2022 at 2:30AM",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := humanDate(tc.time)
			if tc.want != got {
				t.Errorf("want: %v, got %v", tc.want, got)
			}
		})
	}
}

func TestNewTemplateCache(t *testing.T) {
	tests := map[string]struct {
		files         fstest.MapFS
		wantTemplates map[string]bool
		wantError     bool
	}{
		"valid template files": {
			files: fstest.MapFS{
				"html/layout.tmpl":          &fstest.MapFile{},
				"html/partials/header.tmpl": &fstest.MapFile{},
				"html/pages/index.tmpl":     &fstest.MapFile{},
			},
			wantTemplates: map[string]bool{"index.tmpl": true},
			wantError:     false,
		},
		"invalid template file": {
			files: fstest.MapFS{
				"html/layout.tmpl":          &fstest.MapFile{},
				"html/partials/header.tmpl": &fstest.MapFile{},
				"html/pages/invalid.tmpl": &fstest.MapFile{
					Data: []byte("{{if}}"),
				},
			},
			wantTemplates: map[string]bool{},
			wantError:     true,
		},
		"missing required template": {
			files: fstest.MapFS{
				"html/partials/header.tmpl": &fstest.MapFile{},
				"html/pages/index.tmpl":     &fstest.MapFile{},
			},
			wantTemplates: map[string]bool{},
			wantError:     true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cache, err := newTemplateCache(tc.files)

			if err == nil && tc.wantError || err != nil && !tc.wantError {
				t.Errorf("want error: %t, got: %v", tc.wantError, err)
			}

			for wantTemplate := range tc.wantTemplates {
				_, ok := cache[wantTemplate]
				if !ok {
					t.Errorf("template not found in cache: %s", wantTemplate)
				}
			}

			for template := range cache {
				ok := tc.wantTemplates[template]
				if !ok {
					t.Errorf("unexpected template found in cache: %s", template)
				}
			}
		})
	}
}

func TestNewTemplateData(t *testing.T) {
	tests := map[string]struct {
		appEnv  string
		urlPath string
		flash   string
		want    *templateData
	}{
		"production": {
			appEnv:  "production",
			urlPath: "/",
			flash:   "",
			want: &templateData{
				Flash:           "",
				PublicPath:      staticURL,
				StaticURL:       staticURL,
				URLPath:         "/",
				AppEnv:          "production",
				IsAuthenticated: false,
			},
		},
		"dev": {
			appEnv:  "dev",
			urlPath: "/login",
			flash:   "login successful",
			want: &templateData{
				Flash:           "login successful",
				PublicPath:      "/public",
				StaticURL:       staticURL,
				URLPath:         "/login",
				AppEnv:          "dev",
				IsAuthenticated: false,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			os.Setenv("APP_ENV", tc.appEnv)
			app := &application{
				sessionManager: scs.New(),
			}

			ctx, _ := app.sessionManager.Load(context.Background(), "")
			app.sessionManager.Put(ctx, "flash", tc.flash)
			r, _ := http.NewRequestWithContext(ctx, "GET", tc.urlPath, nil)

			templateData := app.newTemplateData(r)

			if templateData.Flash != tc.want.Flash {
				t.Errorf("Flash wrong, want: %s, got %s", tc.want.Flash, templateData.Flash)
			}
			if templateData.PublicPath != tc.want.PublicPath {
				t.Errorf("PublicPath wrong, want: %s, got %s", tc.want.PublicPath, templateData.PublicPath)
			}
			if templateData.StaticURL != tc.want.StaticURL {
				t.Errorf("StaticURL wrong, want: %s, got %s", tc.want.StaticURL, templateData.StaticURL)
			}
			if templateData.URLPath != tc.want.URLPath {
				t.Errorf("URLPath wrong, want: %s, got %s", tc.want.URLPath, templateData.URLPath)
			}
			if templateData.AppEnv != tc.want.AppEnv {
				t.Errorf("AppEnv wrong, want: %s, got %s", tc.want.AppEnv, templateData.AppEnv)
			}
			if templateData.IsAuthenticated != tc.want.IsAuthenticated {
				t.Errorf("IsAuthenticated wrong, want: %t, got %t", tc.want.IsAuthenticated, templateData.IsAuthenticated)
			}
		})
	}
}

func TestRenderTemplate(t *testing.T) {
	files := fstest.MapFS{
		"html/layout.tmpl": &fstest.MapFile{
			Data: []byte("{{define \"layout\"}}{{end}}"),
		},
		"html/pages/home.tmpl": &fstest.MapFile{},
	}

	app := application{
		templateCache: map[string]*template.Template{
			"home": template.Must(template.New("home").ParseFS(files, "html/layout.tmpl", "html/pages/home.tmpl")),
		},
		errorLog: log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile),
	}

	tests := map[string]struct {
		statusCode     int
		page           string
		data           *templateData
		wantStatusCode int
	}{
		"valid template and data": {
			statusCode:     http.StatusOK,
			page:           "home",
			data:           &templateData{PageData: "hello world"},
			wantStatusCode: http.StatusOK,
		},
		"template doesn't exist": {
			statusCode:     http.StatusOK,
			page:           "dne",
			data:           &templateData{PageData: "dne"},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()

			app.renderTemplate(w, tc.statusCode, tc.page, tc.data)

			if w.Code != tc.wantStatusCode {
				t.Errorf("unexpected statusCode, want: %d, got: %d", tc.wantStatusCode, w.Code)
			}
		})
	}
}
