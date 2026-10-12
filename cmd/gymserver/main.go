// Command gymserver keeps the 5/3/1 program on an always-on machine (a
// Raspberry Pi) so it can be read and logged from a phone at the gym. It
// serves the workouts JSON API (workouts.NewHandler), which the desktop app
// also uses once its workout server URL is set, plus a small phone page
// built from frontend/gym (see scripts/build-gym.sh).
//
// It has no login of its own: keep it on the home network and reach it
// from outside through Tailscale.
package main

import (
	"embed"
	"flag"
	"io/fs"
	"log"
	"net/http"

	"focusup/internal/workouts"
)

//go:embed all:dist
var dist embed.FS

func main() {
	addr := flag.String("addr", ":8787", "address to listen on")
	data := flag.String("data", "", "workouts JSON file (default: the user config dir, like the desktop app)")
	flag.Parse()

	var (
		service *workouts.Service
		err     error
	)
	if *data == "" {
		service, err = workouts.NewService()
	} else {
		service, err = workouts.NewServiceAt(*data)
	}
	if err != nil {
		log.Fatalf("failed to load workouts: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/api/", workouts.NewHandler(service))
	mux.Handle("/", pageHandler())

	log.Printf("gymserver listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}

// pageHandler serves the built phone page from dist/app.
func pageHandler() http.Handler {
	page, err := fs.Sub(dist, "dist/app")
	if err != nil {
		log.Fatal(err)
	}
	if _, err := fs.Stat(page, "index.html"); err != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "phone page not built: run scripts/build-gym.sh", http.StatusNotFound)
		})
	}
	return http.FileServerFS(page)
}
