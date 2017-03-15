package server

import (
	"encoding/gob"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/strava/go.strava"
	"go.roman.zone/bici/storage"
	"go.roman.zone/bici/strava/activity"
	"go.roman.zone/bici/strava/segment"
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
	r.HandleFunc(path, stravaAuth.HandlerFunc(authSuccessHandler, authFailureHandler))

	const staticPathPrefix = "/static"
	r.PathPrefix(staticPathPrefix).Handler(http.StripPrefix(staticPathPrefix, http.FileServer(http.Dir(staticFilesLoc))))

	return r
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	check(renderTemplate("index", w, r, Page{}))
}

func athleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := atoi64(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	athlete, err := storage.GetAthlete(id)
	if err != nil {
		// FIXME: This might not be a correct status to return (if athlete not found)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	summitActivities, err := storage.GetSummitActivities(athlete.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	check(renderTemplate("profile", w, r, Page{
		Title: "Athlete",
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
	summits, err:=storage.GetAllSummits()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	check(renderTemplate("summit_list", w, r, Page{
		Data: &summits,
	}))
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

	check(renderTemplate("summit_details", w,r, Page{
		Title: "Summit details",
		Data: summit,
	}))
}

func aboutPageHandler(w http.ResponseWriter, r *http.Request) {
	check(renderTemplate("about", w,r, Page{
		Title: "About",
	}))
}
