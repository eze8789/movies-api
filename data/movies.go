package data

import (
	"time"

	"github.com/eze8789/movies-api/validator"
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

type Movie struct {
	ID        int64     `json:"id,omitempty"`
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title,omitempty"`
	Year      int32     `json:"year,omitempty"`
	Runtime   Runtime   `json:"runtime,omitempty"`
	Genres    []string  `json:"genres,omitempty"`
	Version   int32     `json:"version,omitempty"`
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
