package main

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"

	"github.com/0xhjohnson/clacksy/models"
	"github.com/0xhjohnson/clacksy/validator"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.renderTemplate(w, http.StatusOK, "home.tmpl", data)
}

type signupForm struct {
	Email    string
	Password string
	validator.Validator
}

func (app *application) newUserForm(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = signupForm{}
	app.renderTemplate(w, http.StatusOK, "signup.tmpl", data)
}

func (app *application) addNewUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := signupForm{
		Email:    r.PostForm.Get("email"),
		Password: r.PostForm.Get("password"),
	}

	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannnot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRegex), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannnot be blank")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.renderTemplate(w, http.StatusUnprocessableEntity, "signup.tmpl", data)
		return
	}

	err = app.users.Insert(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.AddFieldError("email", "Email address is already in use")

			data := app.newTemplateData(r)
			data.Form = form
			app.renderTemplate(w, http.StatusUnprocessableEntity, "signup.tmpl", data)
		} else {
			app.serverError(w, err)
		}

		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Your signup was successful. Please log in.")

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

type loginForm struct {
	Email    string
	Password string
	validator.Validator
}

func (app *application) loginUserForm(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = loginForm{}
	app.renderTemplate(w, http.StatusOK, "login.tmpl", data)
}

func (app *application) loginUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := loginForm{
		Email:    r.PostForm.Get("email"),
		Password: r.PostForm.Get("password"),
	}

	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannnot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRegex), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannnot be blank")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.renderTemplate(w, http.StatusUnprocessableEntity, "login.tmpl", data)
		return
	}

	userID, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddNonFieldError("Email or password is incorrect")

			data := app.newTemplateData(r)
			data.Form = form
			app.renderTemplate(w, http.StatusUnprocessableEntity, "login.tmpl", data)
		} else {
			app.serverError(w, err)
		}

		return
	}

	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "authenticatedUserID", userID.String())

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) logoutUser(w http.ResponseWriter, r *http.Request) {
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Remove(r.Context(), "authenticatedUserID")
	app.sessionManager.Put(r.Context(), "flash", "You've been logged out successfully")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type soundtestForm struct {
	Keyboard       string
	Keyswitch      string
	PlateMaterial  string
	KeycapMaterial string
	Parts          models.AllParts
	validator.Validator
}

func (app *application) addSoundtestForm(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	keebParts, err := app.parts.GetAll()
	if err != nil {
		app.serverError(w, err)
		return
	}

	data.Form = soundtestForm{
		Parts: models.AllParts{
			Keyboards:       keebParts.Keyboards,
			Switches:        keebParts.Switches,
			PlateMaterials:  keebParts.PlateMaterials,
			KeycapMaterials: keebParts.KeycapMaterials,
		},
	}
	app.renderTemplate(w, http.StatusOK, "soundtest.tmpl", data)
}

func (app *application) addSoundtest(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("soundtest")
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	defer file.Close()

	keebParts, err := app.parts.GetAll()
	if err != nil {
		app.serverError(w, err)
		return
	}

	form := soundtestForm{
		Keyboard:       r.PostForm.Get("keyboard"),
		Keyswitch:      r.PostForm.Get("keyswitch"),
		PlateMaterial:  r.PostForm.Get("plate-material"),
		KeycapMaterial: r.PostForm.Get("keycap-material"),
		Parts: models.AllParts{
			Keyboards:       keebParts.Keyboards,
			Switches:        keebParts.Switches,
			PlateMaterials:  keebParts.PlateMaterials,
			KeycapMaterials: keebParts.KeycapMaterials,
		},
	}

	form.CheckField(validator.NotBlank(form.Keyboard), "keyboard", "This field cannnot be blank")
	form.CheckField(validator.NotBlank(form.Keyswitch), "keyswitch", "This field cannnot be blank")
	form.CheckField(validator.NotBlank(form.PlateMaterial), "plate-material", "This field cannnot be blank")
	form.CheckField(validator.NotBlank(form.KeycapMaterial), "keycap-material", "This field cannnot be blank")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.renderTemplate(w, http.StatusUnprocessableEntity, "soundtest.tmpl", data)
		return
	}

	userID := app.sessionManager.GetString(r.Context(), "authenticatedUserID")
	objKey := filepath.Join("soundtests", userID, fileHeader.Filename)

	_, err = app.s3Client.PutObject(&s3.PutObjectInput{
		Body:   file,
		Bucket: aws.String(os.Getenv("B2_BUCKET")),
		Key:    aws.String(objKey),
	})
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = app.soundtests.Insert(objKey, form.Keyboard, form.PlateMaterial, form.KeycapMaterial, form.Keyswitch, userID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Your soundtest was added successfully")

	http.Redirect(w, r, "/soundtest/new", http.StatusSeeOther)
}
