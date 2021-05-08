package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/eze8789/movies-api/validator"
	"github.com/lib/pq"
)

// {
//     "id": 123,
//     "title": "Casablanca",
//     "runtime": 102,
//     "genres": [
//         "drama",
//         "romance",
//         "war"
//     ],
//     "version": 1
// }

const QueryTimeOut = 3

type Movie struct {
	ID        int64     `json:"id,omitempty"`
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title,omitempty"`
	Year      int32     `json:"year,omitempty"`
	Runtime   Runtime   `json:"runtime,omitempty"`
	Genres    []string  `json:"genres,omitempty"`
	Version   int32     `json:"version,omitempty"`
}

type MovieModel struct {
	*sql.DB
}

func ValidateMovie(v *validator.Validator, movie *Movie) {
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must be less or equal current year")
	v.Check(movie.Year >= 1900, "year", "must be after 1900")

	v.Check(len(movie.Genres) > 0, "genres", "add at least one genre")
	v.Check(v.Unique(movie.Genres), "genres", "genres must be unique")

	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")
}

func (m *MovieModel) Insert(movie *Movie) error {
	stmt := `
	INSERT INTO movies (title, year, runtime, genres) 
	VALUES ($1, $2, $3, $4)
	RETURNING id, created_at, version`
	args := []interface{}{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}
	ctx, cancel := context.WithTimeout(context.Background(), QueryTimeOut*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, stmt, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

func (m *MovieModel) Update(movie *Movie) error {
	stmt := `UPDATE movies
	SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
	WHERE id = $5 AND version =$6
	RETURNING version`
	args := []interface{}{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres), movie.ID, movie.Version}

	ctx, cancel := context.WithTimeout(context.Background(), QueryTimeOut*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, stmt, args...).Scan(&movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

func (m *MovieModel) Get(id int64) (*Movie, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	stmt := `SELECT id, created_at, title, year, runtime, genres, version
	FROM movies
	WHERE id = $1`

	var movie Movie

	ctx, cancel := context.WithTimeout(context.Background(), QueryTimeOut*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, stmt, id).Scan(&movie.ID, &movie.CreatedAt, &movie.Title,
		&movie.Year, &movie.Runtime, pq.Array(&movie.Genres), &movie.Version)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &movie, nil
}

func (m *MovieModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	stmt := `DELETE FROM movies WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), QueryTimeOut*time.Second)
	defer cancel()

	r, err := m.DB.ExecContext(ctx, stmt, id)
	if err != nil {
		return err
	}

	rows, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrRecordNotFound
	}
	return nil
}
