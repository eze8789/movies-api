package data

import (
	"database/sql"
	"errors"
)

const QueryTimeOut = 3

var (
	ErrRecordNotFound  = errors.New("record not found")
	ErrEditConflict    = errors.New("edit conflict, please try again")
	ErrDuplicatedEmail = errors.New("email already registered")
)

type Models struct {
	Movies      MovieModel
	Users       UserModel
	Tokens      TokensModel
	Permissions PermissionsModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Movies:      MovieModel{DB: db},
		Users:       UserModel{DB: db},
		Tokens:      TokensModel{DB: db},
		Permissions: PermissionsModel{DB: db},
	}
}
