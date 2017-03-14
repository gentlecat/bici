package server

import (
	"encoding/gob"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/strava/go.strava"
	"go.roman.zone/turbo-parakeet/storage"
	"go.roman.zone/turbo-parakeet/strava/activity"
	"go.roman.zone/turbo-parakeet/strava/segment"
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

type Page struct {
	Title string
	User  *strava.AthleteDetailed
	Data  interface{}
}

func StartServer() {
	log.Println("Initializing...")
	initAll()

	clientIDString := os.Getenv("STRAVA_CLIENT_ID")
	if clientIDString == "" {
		log.Fatal("STRAVA_CLIENT_ID env variable is missing")
	}
	clientID, err := strconv.Atoi(clientIDString)
	check(err)
	strava.ClientId = clientID
	strava.ClientSecret = os.Getenv("STRAVA_CLIENT_SECRET")
	stravaAuth = &strava.OAuthAuthenticator{
		CallbackURL: fmt.Sprintf("%s://%s:%d/oauth",
			os.Getenv("OAUTH_CALLBACK_PROTOCOL"), os.Getenv("OAUTH_CALLBACK_HOST"), *listenPort),
		RequestClientGenerator: nil,
	}

	listenAddr := fmt.Sprintf("%s:%d", *listenHost, *listenPort)
	log.Printf("Starting server on %s...\n", listenAddr)
	check(http.ListenAndServe(listenAddr, makeRouter()))
}

func initAll() {
	flag.Parse()
	templLoc = filepath.Join(*resourcesLoc, "templates")
	staticFilesLoc = filepath.Join(*resourcesLoc, "static")

	renderTemplates(templLoc)

	gob.Register(&UserSession{})
}

func makeRouter() *mux.Router {
	r := mux.NewRouter().StrictSlash(true)

	r.HandleFunc("/", indexHandler)
	r.HandleFunc("/about", aboutPageHandler)
	r.HandleFunc("/athletes/{id:[0-9]+}", athleteHandler)
	r.HandleFunc("/activities/{id:[0-9]+}", activityHandler)
	r.HandleFunc("/summits", summitsHandler)
	r.HandleFunc("/summits/{id:[0-9]+}", summitDetailsHandler)

	// Strava OAuth
	r.HandleFunc("/login", loginHandler)
	path, err := stravaAuth.CallbackPath()
	check(err)
	r.HandleFunc(path, stravaAuth.HandlerFunc(authSuccess, authFailure))

	const staticPathPrefix = "/static"
	r.PathPrefix(staticPathPrefix).Handler(http.StripPrefix(staticPathPrefix, http.FileServer(http.Dir(staticFilesLoc))))

	return r
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	check(renderTemplate("index", w, Page{}))
}

func athleteHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, SESSION_NAME)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	userSessionInterf, ok := session.Values[SESSION_KEY_USER]
	if !ok {
		http.Error(w, "Can't find session info", http.StatusInternalServerError)
		return
	}
	userSession, ok := userSessionInterf.(*UserSession)
	if !ok {
		http.Error(w, "Can't type cast user session", http.StatusInternalServerError)
		return
	}

	loggedInUser, err := storage.GetAthlete(userSession.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	id, err := atoi64(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	athlete, err := storage.GetAthlete(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	summitActivities, err := storage.GetSummitActivities(athlete.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	check(renderTemplate("profile", w, Page{
		User: &loggedInUser,
		Data: struct {
			Athlete          *strava.AthleteDetailed
			SummitActivities *[]*storage.SummitEffortActivity
		}{
			&athlete,
			&summitActivities,
		},
	}))
}

func activityHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Shouldn't retrieve activities here!
	vars := mux.Vars(r)
	id, err := atoi64(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	a, err := activity.GetActivity(strava.NewClient(access_token), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := storage.SaveActivity(a); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	has := segment.HasAttemptedAny(a, []int64{12083951})

	content, _ := json.MarshalIndent(has, "", " ")
	fmt.Fprint(w, string(content))

	//err = renderTemplate("profile", w, Page{ })
}

func summitsHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: List of tracked summits
}

func summitDetailsHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Show info about the summit, athletes who completed it
	vars := mux.Vars(r)
	id, err := atoi64(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	summit, err := storage.GetSummit(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	check(renderTemplate("summit_details", w, Page{
		Data: summit,
	}))
}

func aboutPageHandler(w http.ResponseWriter, r *http.Request) {
	check(renderTemplate("about", w, Page{}))
}
