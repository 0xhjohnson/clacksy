package main

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/0xhjohnson/clacksy/models"
	"github.com/0xhjohnson/clacksy/validator"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-chi/chi/v5"
)

const (
	MB = 1 << 20
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

	http.Redirect(w, r, "/vote", http.StatusSeeOther)
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
	r.Body = http.MaxBytesReader(w, r.Body, 24*MB)
	err := r.ParseMultipartForm(24 * MB)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	defer r.MultipartForm.RemoveAll()

	file, mpFileHeader, err := r.FormFile("soundtest")
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	defer file.Close()

	buffer := bytes.NewBuffer(make([]byte, 0, mpFileHeader.Size))
	_, err = io.Copy(buffer, file)
	if err != nil {
		app.serverError(w, err)
		return
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		app.serverError(w, err)
		return
	}

	fileHeader := make([]byte, 512)
	_, err = file.Read(fileHeader)
	if err != nil {
		app.serverError(w, err)
		return
	}

	filetype := http.DetectContentType(fileHeader)
	isValidFileType := strings.Contains(filetype, "audio") || strings.Contains(filetype, "video")
	if !isValidFileType {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	var filename = mpFileHeader.Filename

	if strings.Contains(filetype, "video") {
		tmpFile, err := os.CreateTemp("", "soundtest-*"+filepath.Ext(filename))
		if err != nil {
			app.serverError(w, err)
			return
		}
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		_, err = tmpFile.Write(buffer.Bytes())
		if err != nil {
			app.serverError(w, err)
			return
		}

		args := []string{"-i", tmpFile.Name(), "-t", "40", "-vn", "-acodec", "copy", "-f", "adts", "pipe:1"}
		out, err := exec.Command("ffmpeg", args...).Output()
		if err != nil {
			app.serverError(w, err)
			return
		}

		filename = filenameWithoutExt(mpFileHeader.Filename) + ".aac"
		buffer.Reset()
		buffer.Write(out)
	}

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
	objKey := filepath.Join("soundtests", userID, filename)

	reader := bytes.NewReader(buffer.Bytes())

	_, err = app.s3Client.PutObject(&s3.PutObjectInput{
		Body:   reader,
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

type votePageData struct {
	SoundTests []models.SoundTestVote
	HasMore    bool
	Page       int
	PrevPage   int
	NextPage   int
}

func (app *application) vote(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	page := r.Context().Value(pageContextKey).(int)
	perPage := 10
	userID := app.sessionManager.GetString(r.Context(), "authenticatedUserID")

	soundtests, err := app.soundtests.GetLatest(page, perPage, userID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data.PageData = votePageData{
		SoundTests: soundtests,
		HasMore:    len(soundtests) > 0 && soundtests[0].TotalTests > len(soundtests)*(page+1),
		Page:       page,
		PrevPage:   page - 1,
		NextPage:   page + 1,
	}

	app.renderTemplate(w, http.StatusOK, "vote.tmpl", data)
}

func (app *application) upvote(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	page := r.Context().Value(pageContextKey).(int)
	perPage := 10
	userID := app.sessionManager.GetString(r.Context(), "authenticatedUserID")
	soundtestID := chi.URLParam(r, "soundtestID")
	prevVote := r.FormValue("previous-vote")

	vote := 1
	if prevVote == "1" {
		vote = 0
	}

	err := app.votes.Upsert(soundtestID, vote, userID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	soundtests, err := app.soundtests.GetLatest(page, perPage, userID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data.PageData = votePageData{
		SoundTests: soundtests,
		HasMore:    true,
		Page:       page,
		PrevPage:   page - 1,
		NextPage:   page + 1,
	}

	app.renderTemplate(w, http.StatusOK, "vote.tmpl", data)
}

func (app *application) downvote(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	page := r.Context().Value(pageContextKey).(int)
	perPage := 10
	userID := app.sessionManager.GetString(r.Context(), "authenticatedUserID")
	soundtestID := chi.URLParam(r, "soundtestID")
	prevVote := r.FormValue("previous-vote")

	vote := -1
	if prevVote == "-1" {
		vote = 0
	}

	err := app.votes.Upsert(soundtestID, vote, userID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	soundtests, err := app.soundtests.GetLatest(page, perPage, userID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data.PageData = votePageData{
		SoundTests: soundtests,
		HasMore:    true,
		Page:       page,
		PrevPage:   page - 1,
		NextPage:   page + 1,
	}

	app.renderTemplate(w, http.StatusOK, "vote.tmpl", data)
}

type dailySound struct {
	SoundTest      models.SoundTest
	Parts          models.AllParts
	Keyboard       string
	Keyswitch      string
	PlateMaterial  string
	KeycapMaterial string
}

func (app *application) getDailySound(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	soundtest, err := app.soundtests.GetDaily()
	if err != nil {
		app.serverError(w, err)
		return
	}
	partOpts, err := app.parts.GetDaily(soundtest.KeyboardID, soundtest.KeyswitchID, soundtest.PlateMaterialID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data.PageData = dailySound{
		SoundTest: soundtest,
		Parts:     partOpts,
	}

	app.renderTemplate(w, http.StatusOK, "play.tmpl", data)
}

type dailyPlayForm struct {
	Keyboard       string
	Keyswitch      string
	PlateMaterial  string
	KeycapMaterial string
	validator.Validator
}

func (app *application) addPlayResult(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	soundtest, err := app.soundtests.GetDaily()
	if err != nil {
		app.serverError(w, err)
		return
	}
	partOpts, err := app.parts.GetDaily(soundtest.KeyboardID, soundtest.KeyswitchID, soundtest.PlateMaterialID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data.PageData = dailySound{
		SoundTest: soundtest,
		Parts:     partOpts,
	}

	err = r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := dailyPlayForm{
		Keyboard:       r.PostForm.Get("keyboard"),
		Keyswitch:      r.PostForm.Get("keyswitch"),
		PlateMaterial:  r.PostForm.Get("plate-material"),
		KeycapMaterial: r.PostForm.Get("keycap-material"),
	}

	form.CheckField(validator.NotBlank(form.Keyboard), "keyboard", "This field cannnot be blank")
	form.CheckField(validator.NotBlank(form.Keyswitch), "keyswitch", "This field cannnot be blank")
	form.CheckField(validator.NotBlank(form.PlateMaterial), "plate-material", "This field cannnot be blank")
	form.CheckField(validator.NotBlank(form.KeycapMaterial), "keycap-material", "This field cannnot be blank")

	data.Form = form

	if !form.Valid() {
		app.renderTemplate(w, http.StatusUnprocessableEntity, "soundtest.tmpl", data)
		return
	}

	userID := app.sessionManager.GetString(r.Context(), "authenticatedUserID")

	err = app.soundtests.AddPlay(soundtest.ID, userID, form.Keyboard, form.PlateMaterial, form.KeycapMaterial, form.Keyswitch)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.renderTemplate(w, http.StatusOK, "grade.tmpl", data)
}

func (app *application) getGrade(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	playResult := r.Context().Value(userPlayContextKey).(models.SoundTestPlay)
	data.PageData = playResult

	app.renderTemplate(w, http.StatusOK, "grade.tmpl", data)
}

type profileForm struct {
	Name     string
	Username string
	Email    string
	validator.Validator
}

func (app *application) getUserProfile(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	userID := app.sessionManager.GetString(r.Context(), "authenticatedUserID")

	profile, err := app.users.GetProfileInfo(userID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data.Form = profileForm{
		Name:     profile.Name,
		Username: profile.Username,
		Email:    profile.Email,
	}

	app.renderTemplate(w, http.StatusOK, "profile.tmpl", data)
}

func (app *application) updateUserProfile(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := profileForm{
		Name:     r.PostForm.Get("name"),
		Username: r.PostForm.Get("username"),
		Email:    r.PostForm.Get("email"),
	}

	form.CheckField(validator.MinChars(form.Username, 3), "username", "This field must be at least 3 characters.")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannnot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRegex), "email", "This field must be a valid email address")

	data.Form = form

	if !form.Valid() {
		app.renderTemplate(w, http.StatusUnprocessableEntity, "profile.tmpl", data)
		return
	}

	userID := app.sessionManager.GetString(r.Context(), "authenticatedUserID")

	err = app.users.UpdateProfile(userID, form.Email, form.Name, form.Username)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateUsername) {
			form.AddFieldError("username", "Username is already in use")

			data := app.newTemplateData(r)
			data.Form = form
			app.renderTemplate(w, http.StatusUnprocessableEntity, "profile.tmpl", data)
		} else {
			app.serverError(w, err)
		}

		return
	}

	profile, err := app.users.GetProfileInfo(userID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data.Form = profileForm{
		Name:     profile.Name,
		Username: profile.Username,
		Email:    profile.Email,
	}

	app.renderTemplate(w, http.StatusOK, "profile.tmpl", data)
}
