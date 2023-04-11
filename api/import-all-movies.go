package handler

import (
	"fmt"
	"github.com/lsmoura/omdb-api/database"
	"github.com/lsmoura/omdb-api/logging"
	"github.com/lsmoura/omdb-api/omdb"
	"net/http"
	"os"
)

func APIImportAllMovies(w http.ResponseWriter, r *http.Request) {
	// protect the endpoint with a secret
	requiredSecret := os.Getenv("OMDB_SECRET")
	if requiredSecret == "" {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error: OMDB_SECRET not set")
		return
	}

	auth := r.URL.Query().Get("auth")
	if auth != requiredSecret {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Error: unauthorized")
		return
	}

	logging.LoggerMiddleware(http.HandlerFunc(importAllMoviesHandler), nil).ServeHTTP(w, r)
}

func importAllMoviesHandler(w http.ResponseWriter, r *http.Request) {
	db, err := database.DB()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error: %s", err)
		return
	}
	defer db.Close()

	if err := omdb.ImportAllMovies(r.Context(), db); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error: %s", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}
