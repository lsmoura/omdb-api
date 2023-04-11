package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

func DB() (*sql.DB, error) {
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

func GetMovies(db *sql.DB) ([]Movie, error) {
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

func GetMovieWithID(db *sql.DB, id int64) (*Movie, error) {
	row := db.QueryRow("SELECT id, name, parent_id, date FROM movies WHERE id = $1", id)

	var movie Movie
	if err := row.Scan(&movie.ID, &movie.Name, &movie.ParentID, &movie.Date); err != nil {
		return nil, fmt.Errorf("row.Scan: %w", err)
	}

	return &movie, nil
}

func GetMovieForIMDBID(db *sql.DB, imdbID string) (*Movie, error) {
	row := db.QueryRow("SELECT movie_id FROM movie_links WHERE source = 'imdbmovie' AND key = $1", imdbID)

	var movieID int64
	if err := row.Scan(&movieID); err != nil {
		return nil, fmt.Errorf("row.Scan: %w", err)
	}

	return GetMovieWithID(db, movieID)
}
