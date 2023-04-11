package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/lsmoura/omdb-api/database"
	"github.com/lsmoura/omdb-api/logging"
	"net/http"
	"time"
)

// Handler is the HTTP handler for the API, handled by the lambda
var Handler http.HandlerFunc = logging.LoggerMiddleware(http.HandlerFunc(handler), nil).ServeHTTP

func handler(w http.ResponseWriter, r *http.Request) {
	db, err := database.DB()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error: %s", err)
		return
	}
	defer db.Close()

	movies, err := database.GetMovies(db)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error: %s", err)
		return
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(movies); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error: %s", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Expires", time.Now().Add(time.Hour).Format(time.RFC1123))
	if _, err := w.Write(buf.Bytes()); err != nil {
		if logger := logging.LoggerFromContext(r.Context()); logger != nil {
			logger.Error("w.Write", "error", err)
		}
	}
}
