package validator

import "regexp"

var (
	EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$") //nolint:lll
)

type Validator struct {
	Errors map[string]string
}

func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

func (v *Validator) AddError(key, message string) {
	if _, ok := v.Errors[key]; !ok {
		v.Errors[key] = message
	}
}

func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

func (v *Validator) In(val string, list ...string) bool {
	for i := range list {
		if val == list[i] {
			return true
		}
	}
	return false
}

func Matches(val string, rx *regexp.Regexp) bool {
	return rx.MatchString(val)
}

//nolint:gocritic
func (v *Validator) Unique(values []string) bool {
	uniqueVals := make(map[string]bool)
	for _, v := range values {
		if _, ok := uniqueVals[v]; !ok {
			uniqueVals[v] = true
			continue
		}
		if _, ok := uniqueVals[v]; ok {
			uniqueVals[v] = false
		}
	}
	for _, v := range uniqueVals {
		if !v {
			return false
		}
	}
	return true
	// for _, v := range values {
	// 	uniqueVals[v] = true
	// }
	// return len(values) == len(uniqueVals)
}
