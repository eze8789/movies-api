package data

import (
	"database/sql"
	"errors"
)

const QueryTimeOut = 3

var (
	ErrRecordNotFound  = errors.New("record not found")
	ErrEditConflict    = errors.New("edit conflict")
	ErrDuplicatedEmail = errors.New("email already registered")
)

type Models struct {
	Movies MovieModel
	Users  UserModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{DB: db},
		Users:  UserModel{DB: db},
	}
}
