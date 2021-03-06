package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/eze8789/movies-api/validator"
	"github.com/julienschmidt/httprouter"
)

func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}
	return id, nil
}

type envelope map[string]interface{}

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	res, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}
	res = append(res, '\n')

	for k, v := range headers {
		w.Header()[k] = v
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(res) //nolint:errcheck
	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	maxBytes := 1_048_576

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("badly-formed JSON at character: %d", syntaxError.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return fmt.Errorf("invalid JSON")
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("incorrect JSON type %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("invalid JSON value at character: %d", unmarshalTypeError.Offset)
		case errors.Is(err, io.EOF):
			return fmt.Errorf("body can not be empty")
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("unknown key found: %s", fieldName)
		case err.Error() == "http: request body too large":
			return fmt.Errorf("payload must not be larger than %d bytes", maxBytes)
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("one single JSON value allowed")
	}

	return nil
}

func GetString(s string) string {
	return os.Getenv(s)
}

func GetInt(s string) (int, error) {
	val := GetString(s)
	n, err := strconv.Atoi(val)
	if err != nil {
		return -1, err
	}
	return n, nil
}

func GetFloat(s string) (float64, error) {
	val := GetString(s)
	n, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return -1, err
	}
	return n, nil
}

func GetBool(s string) bool {
	return GetString(s) == "true"
}

func (app *application) readString(qs url.Values, k, defaultValue string) string {
	s := qs.Get(k)

	if s == "" {
		return defaultValue
	}

	return s
}

func (app *application) readInt(qs url.Values, k string, defaultValue int, v *validator.Validator) int {
	s := qs.Get(k)
	if s == "" {
		return defaultValue
	}

	n, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(k, "must be an integer")
		return defaultValue
	}

	return n
}

func (app *application) readCSV(qs url.Values, k string, defaultValue []string) []string {
	csv := qs.Get(k)

	if csv == "" {
		return defaultValue
	}
	return strings.Split(csv, ",")
}

func (app *application) runBackground(fn func()) {
	app.wg.Add(1)
	go func() {
		defer app.wg.Done()
		// Recover goroutine in case of panic
		defer func() {
			if err := recover(); err != nil {
				app.logger.LogError(fmt.Errorf("%s", err), nil)
			}
		}()
		// Execute background work
		fn()
	}()
}
