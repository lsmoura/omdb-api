package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lsmoura/omdb-api/database"
	"github.com/lsmoura/omdb-api/logging"
	"net/http"
)

type imdbResponse struct {
	ImdbID   string  `json:"imdb_id"`
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	ParentID *int64  `json:"parent_id"`
	Date     *string `json:"date"`
}

func APIImdb(w http.ResponseWriter, r *http.Request) {
	logging.LoggerMiddleware(http.HandlerFunc(imdbHandler), nil).ServeHTTP(w, r)
}

func imdbHandler(w http.ResponseWriter, r *http.Request) {
	db, err := database.DB()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error: %s", err)
		return
	}
	defer db.Close()

	imdbID := r.URL.Query().Get("q")
	if imdbID == "" || imdbID[:2] != "tt" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	movie, err := database.GetMovieForIMDBID(db, imdbID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error: %s", err)
		return
	}

	resp := imdbResponse{
		ImdbID:   imdbID,
		ID:       movie.ID,
		Name:     movie.Name,
		ParentID: movie.ParentID,
		Date:     movie.Date,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error: %s", err)
		return
	}
}
