package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"time"

	"github.com/eze8789/movies-api/validator"
)

const (
	ActivationTokenDuration      = 12 * time.Hour
	AuthenticationTokenDuration  = 4 * time.Hour
	PasswordRecoverTokenDuration = time.Hour
	ScopeActivation              = "activation"
	ScopeAuthentication          = "authentication"
	ScopePasswordReset           = "password-reset"
)

type Token struct {
	PlainToken  string    `json:"token"`
	HashedToken []byte    `json:"-"`
	UserID      int64     `json:"-"`
	Expiry      time.Time `json:"expiry"`
	Scope       string    `json:"-"`
}

type TokensModel struct {
	*sql.DB
}

// genrateToken return a new Token instance with a hashed token with high entropy
func generateToken(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token := &Token{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	// read random bytes form CSPRNG from the OS
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	// token encoded in 32 bytes
	token.PlainToken = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	// generate hash and convert to a slice of bytes
	hash := sha256.Sum256([]byte(token.PlainToken))
	token.HashedToken = hash[:]

	return token, nil
}

// ValidateTokenPlain ensure the token received is within valid values
func ValidateTokenPlain(v *validator.Validator, tokenPlain string) {
	v.Check(tokenPlain != "", "token", "token must be provided")
	v.Check(len(tokenPlain) == 26, "token", "must be equal to 26 bytes long") //nolint:gomnd
}

// Insert write a new Token to the tokens DB
func (tm *TokensModel) Insert(token *Token) error {
	stmt := `INSERT INTO tokens (hash, user_id, expiry, scope)
	VALUES ($1, $2, $3, $4)`
	args := []interface{}{token.HashedToken, token.UserID, token.Expiry, token.Scope}

	ctx, cancel := context.WithTimeout(context.Background(), QueryTimeOut*time.Second)
	defer cancel()

	_, err := tm.DB.ExecContext(ctx, stmt, args...)

	return err
}

// New is wrapper to generate a new Token and store it in the tokens DB using Insert
func (tm *TokensModel) New(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token, err := generateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}

	err = tm.Insert(token)
	return token, err
}

// DeleteAllByUser delete all tokens associated within a user with an specific scope
func (tm *TokensModel) DeleteAllByUser(userID int64, scope string) error {
	stmt := `DELETE FROM tokens
	WHERE user_id = $1 AND scope = $2`
	args := []interface{}{userID, scope}

	ctx, cancel := context.WithTimeout(context.Background(), QueryTimeOut*time.Second)
	defer cancel()

	_, err := tm.DB.ExecContext(ctx, stmt, args...)

	return err
}
