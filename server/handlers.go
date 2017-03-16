package server

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/strava/go.strava"
	"go.roman.zone/cazador/storage"
	"go.roman.zone/cazador/strava/activity"
	"log"
	"net/http"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	topAthletes, err := storage.ListTopAthletes()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	check(renderTemplate("index", w, r, Page{
		Data: topAthletes,
	}))
}

func athleteDetailsHandler(w http.ResponseWriter, r *http.Request) {
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

	http.Error(w, fmt.Sprintf("Not implemented yet. Can't view activity %d. :(", id), http.StatusTeapot)
}

func refreshHandler(w http.ResponseWriter, r *http.Request) {
	isLoggedIn, currentUser, err := GetCurrentSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !isLoggedIn {
		http.Error(w, "You need to be logged in!", http.StatusBadRequest)
	}
	log.Println(fmt.Sprintf("User #%d (%s %s) asked for a refresh.",
		currentUser.Id, currentUser.FirstName, currentUser.LastName))
	accessToken, err := storage.GetAthletesAccessToken(currentUser.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// For now this is using my token
	activity.RetrieveAthlete(accessToken)
	fmt.Fprint(w, "Your activities will be retrieved soon!")
}

func athletesHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Add pagination
	athletes, err := storage.BrowseAthletes(20, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	check(renderTemplate("athlete_list", w, r, Page{
		Data: &athletes,
	}))
}

func summitsHandler(w http.ResponseWriter, r *http.Request) {
	summits, err := storage.GetAllSummits()
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

	check(renderTemplate("summit_details", w, r, Page{
		Title: "Summit details",
		Data:  summit,
	}))
}

func aboutPageHandler(w http.ResponseWriter, r *http.Request) {
	check(renderTemplate("about", w, r, Page{
		Title: "About",
	}))
}
