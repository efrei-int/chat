package main

import (
	"errors"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/facebook"
	"github.com/nicksnyder/go-i18n/i18n"

	_ "github.com/joho/godotenv/autoload"
)

// loadSession loads the session storage and initializes the auth
// providers
func loadSession() {
	store := sessions.NewFilesystemStore(os.TempDir(), []byte(os.Getenv("SESSION_SECRET")))
	store.MaxLength(math.MaxInt64)

	gothic.Store = store

	host := getHost()
	goth.UseProviders(facebook.New(os.Getenv("FACEBOOK_KEY"), os.Getenv("FACEBOOK_SECRET"), host + "/auth/callback?provider=facebook"))
}

// loadLocales loads all i18n strings from the `locales` directory
func loadLocales() {
	files, err := ioutil.ReadDir("locales")
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	for _, file := range files {
		i18n.MustLoadTranslationFile("locales/" + file.Name())
	}
}

// initT returns a new i18n TranslateFunc based on the "Accept-Language"
// header and defaulting to "en"
func initT(acceptLang string, defaultLang string) (T i18n.TranslateFunc) {
	T = i18n.MustTfunc(acceptLang, defaultLang)
	return
}

// getPort returns the port by first looking at any environment variable
// nammed PORT and then defaulting to :8000
func getPort() string {
	if port := os.Getenv("PORT"); port != "" {
		return ":" + port
	}
	return ":8000"
}

// getHost returns the host URL by looking if the app is running in "dev" or
// in production (on Heroku for now)
func getHost() string {
	if env := os.Getenv("ENV"); env == "dev" {
		return "http://localhost:8000"
	}
	return os.Getenv("HOST")
}

// getUser returns the goth.User linked with the current session
// NB: for now we are only using Facebook as an OAuth provider
func getUser(r *http.Request, p string) (goth.User, error) {
	session, _ := gothic.Store.Get(r, p + gothic.SessionName)
	values := session.Values[p]
	if values == nil {
		return goth.User{}, errors.New("cannot find session values")
	}
	
	provider, _ := goth.GetProvider(p)
	sess, _ := provider.UnmarshalSession(values.(string))
	user, err := provider.FetchUser(sess)

	if err != nil {
		return goth.User{}, err
	}

	return user, nil
}
