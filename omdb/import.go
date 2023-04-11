package omdb

import (
	"bufio"
	"compress/bzip2"
	"context"
	"database/sql"
	"fmt"
	"github.com/lsmoura/omdb-api/csv"
	"github.com/lsmoura/omdb-api/logging"
	"net/http"
	"strconv"
	"strings"
)

const (
	AllMoviesURL  = "http://www.omdb.org/data/all_movies.csv.bz2"
	MovieLinksURL = "http://www.omdb.org/data/movie_links.csv.bz2"
)

func scanCSVLine(fileScanner *bufio.Scanner) (string, error) {
	text := fileScanner.Text()
	for len(text) > 0 && (text[len(text)-1] == '\\') {
		if !fileScanner.Scan() {
			return "", fmt.Errorf("unexpected EOF")
		}
		text = text[:len(text)-1] + "\n" + fileScanner.Text()
	}

	return text, nil
}

func tx(db *sql.DB, fn func(tx *sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("db.Begin: %w", err)
	}

	if err := fn(tx); err != nil {
		tx.Rollback()
		return fmt.Errorf("fn: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("tx.Commit: %w", err)
	}

	return nil
}

func allMoviesFieldsToArgs(fields []string) ([]any, error) {
	if len(fields) != 4 {
		return nil, fmt.Errorf("invalid number of fields: %d", len(fields))
	}

	id, err := strconv.Atoi(fields[0])
	if err != nil {
		return nil, fmt.Errorf("strconv.Atoi: %w", err)
	}

	var parentID sql.NullInt64
	if fields[2] != "" && fields[2] != "\\N" {
		parentID.Int64, err = strconv.ParseInt(fields[2], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("strconv.ParseInt: %w", err)
		}
		parentID.Valid = true
	}

	return []any{
		id,
		fields[1],
		parentID,
		fields[3],
	}, nil
}

func ImportAllMovies(ctx context.Context, db *sql.DB) error {
	// download all movies from omdb
	req, err := http.NewRequestWithContext(ctx, "GET", AllMoviesURL, nil)
	if err != nil {
		return fmt.Errorf("http.NewRequest: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http.DefaultClient.Do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http status: %d", resp.StatusCode)
	}

	const sqlPrefix = "INSERT INTO movies (id, name, parent_id, date) VALUES"
	const sqlSuffix = " ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, parent_id = EXCLUDED.parent_id, date = EXCLUDED.date"

	// parse and insert movies in the database
	bz2Reader := bzip2.NewReader(resp.Body)
	lineScanner := bufio.NewScanner(bz2Reader)

	err = injectCSV(ctx, db, lineScanner, sqlPrefix, sqlSuffix, nil, allMoviesFieldsToArgs)

	if err != nil {
		return fmt.Errorf("injectCSV: %w", err)
	}

	return nil
}

func movieLinksFieldsToArgs(fields []string) ([]any, error) {
	if len(fields) != 4 {
		return nil, fmt.Errorf("invalid number of fields: %d", len(fields))
	}

	var movieID int64
	if fields[2] != "" && fields[2] != "\\N" {
		var err error
		movieID, err = strconv.ParseInt(fields[2], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("strconv.ParseInt: %w", err)
		}
	}

	return []any{
		fields[0], // source
		fields[1], // key
		movieID,
		fields[3], // language
	}, nil
}

func injectCSV(ctx context.Context, db *sql.DB, lineScanner *bufio.Scanner, sqlPrefix, sqlSuffix string, prepareFn func(*sql.Tx) error, extractor func([]string) ([]any, error)) error {
	var lines int

	err := tx(db, func(tx *sql.Tx) error {
		if prepareFn != nil {
			if err := prepareFn(tx); err != nil {
				return fmt.Errorf("prepareFn: %w", err)
			}
		}

		// test and skip first line
		if !lineScanner.Scan() {
			return fmt.Errorf("lineScanner.Scan: %w", lineScanner.Err())
		}

		var count int

		var sqlBuf strings.Builder
		var args []any

		for lineScanner.Scan() {
			line, err := scanCSVLine(lineScanner)
			if err != nil {
				return fmt.Errorf("scanCSVLine: %w", err)
			}
			if len(line) == 0 {
				continue
			}

			elements, err := csv.LineSplit(line)
			if err != nil {
				return fmt.Errorf("csv.LineSplit: %w", err)
			}

			lines++

			elementArgs, err := extractor(elements)
			if err != nil {
				return fmt.Errorf("extractor: %w", err)
			}

			if count > 0 {
				sqlBuf.WriteString(", ")
			}

			pieces := make([]string, len(elementArgs))
			for i := range elementArgs {
				count++
				pieces[i] = "$" + strconv.Itoa(count)
			}

			sqlBuf.WriteString(fmt.Sprintf(" (%s)", strings.Join(pieces, ", ")))
			args = append(args, elementArgs...)

			if count >= 4000 {
				fullQuery := sqlPrefix + sqlBuf.String() + sqlSuffix
				if _, err := tx.ExecContext(ctx, fullQuery, args...); err != nil {
					return fmt.Errorf("tx.ExecContext: %w", err)
				}
				count = 0
				sqlBuf.Reset()
				args = nil
			}
		}

		if len(args) > 0 {
			fullQuery := sqlPrefix + sqlBuf.String() + sqlSuffix
			if _, err := tx.ExecContext(ctx, fullQuery, args...); err != nil {
				return fmt.Errorf("tx.ExecContext: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("tx: %w", err)
	}

	return nil
}

func ImportMovieLinks(ctx context.Context, db *sql.DB) error {
	// download all movies from omdb
	req, err := http.NewRequestWithContext(ctx, "GET", MovieLinksURL, nil)
	if err != nil {
		return fmt.Errorf("http.NewRequest: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http.DefaultClient.Do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http status: %d", resp.StatusCode)
	}

	const sqlPrefix = "INSERT INTO movie_links (source, key, movie_id, language_iso_639_1) VALUES"
	const sqlSuffix = " ON CONFLICT DO NOTHING"

	var lines int

	// parse and insert movies in the database
	bz2Reader := bzip2.NewReader(resp.Body)
	lineScanner := bufio.NewScanner(bz2Reader)

	err = injectCSV(ctx, db, lineScanner, sqlPrefix, sqlSuffix, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, "DELETE FROM movie_links;"); err != nil {
			return fmt.Errorf("tx.ExecContext: %w", err)
		}

		if _, err := tx.ExecContext(ctx, "SET CONSTRAINTS ALL DEFERRED;"); err != nil {
			return fmt.Errorf("tx.ExecContext: %w", err)
		}

		return nil
	}, movieLinksFieldsToArgs)

	logging.LoggerFromContext(ctx).Info("ImportMovieLinks", "records", lines)

	if err != nil {
		return fmt.Errorf("injectCSV: %w", err)
	}

	return nil
}
