package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Runtime int32

var ErrInvalidRuntimeFormat = errors.New("invalid runtime format")

func (r Runtime) MarshalJSON() ([]byte, error) {
	val := fmt.Sprintf("%d mins", r)

	quotedVal := strconv.Quote(val)

	return []byte(quotedVal), nil
}

func (r *Runtime) UnmarshalJSON(jsonValue []byte) error {
	unqValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	s := strings.Split(unqValue, " ")
	if len(s) != 2 || s[1] != "mins" {
		return ErrInvalidRuntimeFormat
	}

	n, err := strconv.ParseInt(s[0], 10, 32)
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	*r = Runtime(n)
	return nil
}
