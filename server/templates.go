package server

import (
	"fmt"
	"github.com/strava/go.strava"
	"html/template"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"sync"
)

var (
	templLoc = filepath.Join(*resourcesLoc, "templates")

	templatesMutex sync.Mutex
	templates      map[string]*template.Template

	// TODO: How to make this support multiple types?
	fm = template.FuncMap{
		"divide": func(a, b float64) float64 {
			return a / b
		},
		"subtract": func(a, b float64) float64 {
			return a - b
		},
		"sliceify": func(n uint) []int {
			return make([]int, n)
		},
		"formatDecimal": func(n float64) string {
			return fmt.Sprintf("%.2f", n)
		},
	}
)

type Page struct {
	Title string
	Data  interface{}
}

func renderTemplates(location string) {
	log.Println("Rendering templates...")
	defer log.Println("Done!")
	templatesMutex.Lock()
	defer templatesMutex.Unlock()
	templates = make(map[string]*template.Template)
	templates["index"] = getTemplate(location, "index.html")
	templates["about"] = getTemplate(location, "about.html")
	templates["profile"] = getTemplate(location, "profile.html")
	templates["activity"] = getTemplate(location, "activity.html")
	templates["summit_list"] = getTemplate(location, "summit_list.html")
	templates["summit_details"] = getTemplate(location, "summit_details.html")
}

func getTemplate(location, fileName string) *template.Template {
	return template.Must(template.New(fileName).Funcs(fm).ParseFiles(
		filepath.Join(location, fileName),
		filepath.Join(location, "_base.html"),
	))
}

func renderTemplate(name string, wr io.Writer, r *http.Request, data Page) error {
	templatesMutex.Lock()
	defer templatesMutex.Unlock()
	isLoggedIn, currentUser, err := GetCurrentSession(r)
	if err != nil {
		log.Println(err)
	}
	return templates[name].ExecuteTemplate(wr, "base", struct {
		Page
		IsLoggedIn bool
		User       *strava.AthleteDetailed
	}{
		data,
		isLoggedIn,
		currentUser,
	})
}
