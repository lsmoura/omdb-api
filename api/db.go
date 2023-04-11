package handler

import (
	"database/sql"
	"fmt"
	"os"

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
