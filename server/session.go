package server

import (
	"fmt"
	"github.com/gorilla/sessions"
	"github.com/strava/go.strava"
	"go.roman.zone/bici/storage"
	"net/http"
	"os"
)

const (
	CLUB_ID = 262311

	SESSION_NAME     = "that_session"
	SESSION_KEY_USER = "athlete"
)

var (
	store = sessions.NewCookieStore([]byte(os.Getenv("COOKIE_KEY")))

	stravaAuth *strava.OAuthAuthenticator

	// TODO: REMOVE THIS:
	access_token = os.Getenv("ACCESS_TOKEN")
)

type UserSession struct {
	ID int64
}

func GetCurrentSession(r *http.Request) (isLoggedIn bool, currentUser *strava.AthleteDetailed, err error) {
	session, err := store.Get(r, SESSION_NAME)
	if err != nil {
		// No session
		return false, currentUser, err
	}
	userSessionInterf, ok := session.Values[SESSION_KEY_USER]
	if !ok {
		// Session is there, but correct key isn't
		return false, currentUser, err
	}
	userSession, ok := userSessionInterf.(*UserSession)
	if !ok {
		return false, currentUser, err
	}

	loggedInUser, err := storage.GetAthlete(userSession.ID)
	if err != nil {
		return false, currentUser, err
	}

	return true, &loggedInUser, nil
}

func authSuccessHandler(auth *strava.AuthorizationResponse, w http.ResponseWriter, r *http.Request) {
	// Get a session. We're ignoring the error resulted from decoding an
	// existing session: Get() always returns a session, even if empty.
	session, err := store.Get(r, SESSION_NAME)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set some session values.
	session.Values[SESSION_KEY_USER] = UserSession{
		ID: auth.Athlete.AthleteSummary.AthleteMeta.Id,
	}
	// Save it before we write to the response/return from the handler.
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = storage.SaveLoginData(auth)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("%s://%s/athletes/%d", os.Getenv("OAUTH_CALLBACK_PROTOCOL"), os.Getenv("OAUTH_CALLBACK_HOST"), auth.Athlete.Id), http.StatusTemporaryRedirect)
}

func authFailureHandler(err error, w http.ResponseWriter, r *http.Request) {
	if err == strava.OAuthAuthorizationDeniedErr {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, stravaAuth.AuthorizationURL("pointless_state", strava.Permissions.Public, true),
		http.StatusTemporaryRedirect)
}
