package server

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/rs/cors"
)

var port = ":" + os.Getenv("PORT")
var origin = os.Getenv("ORIGIN")

func Start() {
	mux := http.NewServeMux()
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{origin},
		AllowCredentials: true,
		AllowedMethods:   []string{http.MethodGet},
		AllowedHeaders:   []string{"Token", "Host", "User-Agent", "Accept", "Content-Length", "Content-Type"},
	})

	handler := c.Handler(mux)

	mux.Handle("/", root)
	mux.Handle("/posts", Posts)
	mux.Handle("/authors", Authors)

	log.Fatal(http.ListenAndServe(port, handler))
}

var root = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode("Hi.")
})
