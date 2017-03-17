package server

import (
	"fmt"
	"github.com/gorilla/sessions"
	"github.com/strava/go.strava"
	"go.roman.zone/cazador/storage"
	"go.roman.zone/cazador/strava/activity"
	"net/http"
	"os"
)

const (
	CLUB_ID = 262311

	SESSION_COOKIE_NAME = "session"
	SESSION_KEY_USER    = "athlete"
)

var (
	store = sessions.NewCookieStore([]byte(os.Getenv("COOKIE_KEY")))

	stravaAuth *strava.OAuthAuthenticator
)

type UserSession struct {
	ID int64
}

func GetCurrentSession(r *http.Request) (isLoggedIn bool, currentUser *strava.AthleteDetailed, err error) {
	session, err := store.Get(r, SESSION_COOKIE_NAME)
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

func PurgeCurrentSession(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   SESSION_COOKIE_NAME,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
}

func authSuccessHandler(auth *strava.AuthorizationResponse, w http.ResponseWriter, r *http.Request) {
	// Get a session. We're ignoring the error resulted from decoding an
	// existing session: Get() always returns a session, even if empty.
	session, err := store.Get(r, SESSION_COOKIE_NAME)
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

	// Retrieve all the previous activities for this athlete
	// FIXME: This should be only done once when account is created
	activity.RetrieveAthlete(auth.AccessToken)

	http.Redirect(w, r, fmt.Sprintf("/athletes/%d", auth.Athlete.Id), http.StatusTemporaryRedirect)
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
	isLoggedIn, _, err := GetCurrentSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if isLoggedIn {
		http.Error(w, "You are already logged in", http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, stravaAuth.AuthorizationURL("pointless_state", strava.Permissions.Public, true),
		http.StatusTemporaryRedirect)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	isLoggedIn, _, err := GetCurrentSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !isLoggedIn {
		http.Error(w, "You are not logged in", http.StatusBadRequest)
		return
	}
	PurgeCurrentSession(w, r)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
