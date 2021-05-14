package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"time"

	"github.com/eze8789/movies-api/validator"
	"golang.org/x/crypto/bcrypt"
)

const (
	ErrEmailConstraintPG = `pq: duplicate key value violates unique constraint "users_email_key"`
)

type User struct {
	ID        int64     `json:"id"`
	CreatedAT time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}

type password struct {
	plainPWD  *string
	hashedPWD []byte
}

type UserModel struct {
	*sql.DB
}

func (p *password) Set(s string) error {
	hpwd, err := bcrypt.GenerateFromPassword([]byte(s), 16)
	if err != nil {
		return err
	}

	p.plainPWD = &s
	p.hashedPWD = hpwd

	return nil
}

func (p *password) Match(s string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hashedPWD, []byte(s))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}

// ValidateEmail ensure the email is present and is a valid address
func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email")
}

// ValidatePasswordPlain ensure a password is larger than 8 bytes and less than 72 bytes
// bcrypt truncate anything larger than 72 bytes: https://en.wikipedia.org/wiki/Bcrypt#Maximum_password_length
// for production services this can be pre hashed to avoid this limitation
func ValidatePasswordPlain(v *validator.Validator, pwd string) {
	v.Check(pwd != "", "password", "must be provided")
	v.Check(len(pwd) >= 8, "password", "must be at least 8 bytes long")    //nolint:gomnd
	v.Check(len(pwd) <= 72, "password", "must be less than 72 bytes long") //nolint:gomnd
}

// ValidateUser ensure the User provide a valid name and use ValidateEmail and ValidatePassword
// to ensure those fields are valid
func ValidateUser(v *validator.Validator, u *User) {
	v.Check(u.Name != "", "name", "can not be empty")
	v.Check(len(u.Name) < 100, "name", "must not be larger than 100 bytes") //nolint:gomnd

	ValidateEmail(v, u.Email)

	if u.Password.plainPWD != nil {
		ValidatePasswordPlain(v, *u.Password.plainPWD)
	}

	// no input data problem, adding here as a fallback if some other parts fail
	if u.Password.hashedPWD == nil {
		panic("password hash missed")
	}
}

func (um *UserModel) Insert(user *User) error {
	stmt := `INSERT INTO users (name, email, password_hash, activated)
	VALUES($1, $2, $3, $4)
	RETURNING id, created_at, version`
	args := []interface{}{user.Name, user.Email, user.Password.hashedPWD, user.Activated}

	ctx, cancel := context.WithTimeout(context.Background(), QueryTimeOut*time.Second)
	defer cancel()

	err := um.DB.QueryRowContext(ctx, stmt, args...).Scan(&user.ID, &user.CreatedAT, &user.Version)
	if err != nil {
		switch {
		case err.Error() == ErrEmailConstraintPG:
			return ErrDuplicatedEmail
		default:
			return err
		}
	}
	return nil
}

func (um *UserModel) Update(user *User) error {
	stmt := `UPDATE users
	SET name = $1, email = $2, password_hash = $3, activated = $4, version = version + 1
	WHERE id = $5 AND version = $6
	RETURNING version`
	args := []interface{}{user.Name, user.Email, user.Password.hashedPWD, user.Activated, user.ID, user.Version}

	ctx, cancel := context.WithTimeout(context.Background(), QueryTimeOut*time.Second)
	defer cancel()

	err := um.DB.QueryRowContext(ctx, stmt, args...).Scan(&user.Version)
	if err != nil {
		switch {
		case err.Error() == ErrEmailConstraintPG:
			return ErrDuplicatedEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

func (um *UserModel) GetByEmail(e string) (*User, error) {
	stmt := `SELECT id, created_at, name, email, password_hash, activated, version
	FROM users
	WHERE email = $1`

	var user User
	ctx, cancel := context.WithTimeout(context.Background(), QueryTimeOut*time.Second)
	defer cancel()

	err := um.DB.QueryRowContext(ctx, stmt, e).Scan(&user.ID, &user.CreatedAT, &user.Name,
		&user.Email, &user.Password.hashedPWD, &user.Activated, &user.Version)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrDuplicatedEmail
		default:
			return nil, err
		}
	}
	return &user, nil
}

func (um *UserModel) GetForToken(token, scope string) (*User, error) {
	// calculate hash before compare
	tokenHash := sha256.Sum256([]byte(token))

	stmt := ` SELECT users.id, users.created_at, users.name, users.email, users.password_hash, users.activated, users.version
	FROM users
	INNER JOIN tokens
	ON users.id = tokens.user_id
	WHERE tokens.hash = $1 AND tokens.scope = $2 AND tokens.expiry > $3`
	args := []interface{}{tokenHash[:], scope, time.Now()}

	ctx, cancel := context.WithTimeout(context.Background(), QueryTimeOut*time.Second)
	defer cancel()

	var user User
	err := um.DB.QueryRowContext(ctx, stmt, args...).Scan(&user.ID, &user.CreatedAT, &user.Name, &user.Email,
		&user.Password.hashedPWD, &user.Activated, &user.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}
