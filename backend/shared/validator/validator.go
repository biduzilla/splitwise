package validator

import (
	"fmt"
	"regexp"
	"slices"
)

var (
	EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

type Validator struct {
	Errors map[string]string
}

func New() *Validator {
	return &Validator{
		Errors: make(map[string]string),
	}
}

func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

func (v *Validator) AddError(key, message string) {
	if _, exist := v.Errors[key]; !exist {
		v.Errors[key] = message
	}
}

func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

func In(value string, list ...string) bool {
	return slices.Contains(list, value)
}

func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

func Unique(values []string) bool {
	uniqueValues := make(map[string]bool)

	for _, value := range values {
		uniqueValues[value] = true
	}

	return len(values) == len(uniqueValues)
}

func (v *Validator) CheckStringNotEmpty(value, name string) {
	v.Check(value != "", name, "must be provided")
}

func (v *Validator) CheckStringMinLen(
	value string,
	minLen int,
	name string,
) {
	v.Check(len(value) >= minLen, name, fmt.Sprintf("must be at least %d bytes long", minLen))

}

func (v *Validator) CheckStringMaxLen(
	value string,
	maxLen int,
	name string,
) {
	v.Check(len(value) <= maxLen, "password", fmt.Sprintf("must not be more than %d bytes long", maxLen))
}
