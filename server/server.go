package server

import (
	"encoding/gob"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/strava/go.strava"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"strconv"
)

var (
	listenHost = flag.String("host", "127.0.0.1", "Host to listen on")
	listenPort = flag.Int("port", 8080, "Port to listen on")

	resourcesLoc = flag.String("resources", "res", "Location of templates and static files")
	// TODO: Improve this so that it doesn't have to be reset in initAll after flags are parsed
	staticFilesLoc = filepath.Join(*resourcesLoc, "static")
)

func StartServer() {
	log.Println("Initializing...")
	initAll()

	listenAddr := fmt.Sprintf("%s:%d", *listenHost, *listenPort)
	log.Printf("Starting server on %s...\n", listenAddr)
	check(http.ListenAndServe(listenAddr, makeRouter()))
}

func initAll() {
	flag.Parse()
	templLoc = filepath.Join(*resourcesLoc, "templates")
	staticFilesLoc = filepath.Join(*resourcesLoc, "static")

	renderTemplates(templLoc)

	initOAuth()
}

func initOAuth() {
	gob.Register(&UserSession{})

	clientIDString := os.Getenv("STRAVA_CLIENT_ID")
	if clientIDString == "" {
		log.Fatal("STRAVA_CLIENT_ID env variable is missing")
	}
	clientID, err := strconv.Atoi(clientIDString)
	check(err)
	strava.ClientId = clientID
	strava.ClientSecret = os.Getenv("STRAVA_CLIENT_SECRET")
	callbackURL := fmt.Sprintf("%s://%s/oauth",
		os.Getenv("OAUTH_CALLBACK_PROTOCOL"), os.Getenv("OAUTH_CALLBACK_HOST"))
	if os.Getenv("OAUTH_CALLBACK_HOST") == "localhost" || os.Getenv("OAUTH_CALLBACK_HOST") == "127.0.0.1" {
		callbackURL = fmt.Sprintf("%s://%s:%d/oauth",
			os.Getenv("OAUTH_CALLBACK_PROTOCOL"), os.Getenv("OAUTH_CALLBACK_HOST"), *listenPort)
	}
	stravaAuth = &strava.OAuthAuthenticator{
		CallbackURL:            callbackURL,
		RequestClientGenerator: nil,
	}
}

func makeRouter() *mux.Router {
	r := mux.NewRouter().StrictSlash(true)

	r.HandleFunc("/", indexHandler)
	r.HandleFunc("/about", aboutPageHandler)
	r.HandleFunc("/athletes", athletesHandler)
	r.HandleFunc("/athletes/{id:[0-9]+}", athleteDetailsHandler)
	r.HandleFunc("/activities/{id:[0-9]+}", activityHandler)
	r.HandleFunc("/refresh", refreshHandler)
	r.HandleFunc("/summits", summitsHandler)
	r.HandleFunc("/summits/{id:[0-9]+}", summitDetailsHandler)

	// Strava OAuth
	r.HandleFunc("/login", loginHandler)
	r.HandleFunc("/logout", logoutHandler)
	path, err := stravaAuth.CallbackPath()
	check(err)
	r.HandleFunc(path, stravaAuth.HandlerFunc(authSuccessHandler, authFailureHandler))

	const staticPathPrefix = "/static"
	r.PathPrefix(staticPathPrefix).Handler(http.StripPrefix(staticPathPrefix, http.FileServer(http.Dir(staticFilesLoc))))

	return r
}
