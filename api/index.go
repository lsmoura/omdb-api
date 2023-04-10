package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func getDatabase() (*sql.DB, error) {
	connURL := os.Getenv("DATABASE_URL")
	if connURL == "" {
		return nil, fmt.Errorf("DATABASE_URL not set")
	}

	db, err := sql.Open("postgres", connURL)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}

	return db, nil
}

type Movie struct {
	ID       int64   `json:"id" db:"id"`
	Name     string  `json:"name" db:"name"`
	ParentID *int64  `json:"parent_id" db:"parent_id"`
	Date     *string `json:"date" db:"date"`
}

func getMovies(db *sql.DB) ([]Movie, error) {
	rows, err := db.Query("SELECT id, name, parent_id, date FROM movies")
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	var movies []Movie
	for rows.Next() {
		var movie Movie
		if err := rows.Scan(&movie.ID, &movie.Name, &movie.ParentID, &movie.Date); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		movies = append(movies, movie)
	}

	return movies, nil
}

func Handler(w http.ResponseWriter, r *http.Request) {
	db, err := getDatabase()
	if err != nil {
		fmt.Fprintf(w, "Error: %s", err)
		return
	}
	defer db.Close()

	movies, err := getMovies(db)
	if err != nil {
		fmt.Fprintf(w, "Error: %s", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Expires", time.Now().Add(time.Hour).Format(time.RFC1123))
	if err := json.NewEncoder(w).Encode(movies); err != nil {
		fmt.Fprintf(w, "Error: %s", err)
		return
	}
}
