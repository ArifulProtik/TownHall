// Package validation adapts validator/v10 to the Echo validator interface.
package validation

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/go-playground/validator/v10"
)

var alphanumUnderscoreRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// Error carries validator failures with per-field tags.
type Error struct {
	Err    error
	Fields map[string]string
}

func (e *Error) Error() string { return e.Err.Error() }

// CustomValidator implements echo.Validator via go-playground/validator.
type CustomValidator struct {
	V *validator.Validate
}

// New creates a validator ready for echo.Validator assignment.
func New() *CustomValidator {
	v := validator.New()
	_ = v.RegisterValidation("alphanum_underscore", func(fl validator.FieldLevel) bool {
		return alphanumUnderscoreRegex.MatchString(fl.Field().String())
	})
	// minrunes/maxrunes count characters, not bytes: validator's stock
	// min/max use byte length and reject valid multibyte input.
	_ = v.RegisterValidation("minrunes", func(fl validator.FieldLevel) bool {
		n, err := strconv.Atoi(fl.Param())
		if err != nil {
			return false
		}
		return utf8.RuneCountInString(fl.Field().String()) >= n
	})
	_ = v.RegisterValidation("maxrunes", func(fl validator.FieldLevel) bool {
		n, err := strconv.Atoi(fl.Param())
		if err != nil {
			return false
		}
		return utf8.RuneCountInString(fl.Field().String()) <= n
	})
	return &CustomValidator{V: v}
}

// Validate implements echo.Validator.
func (cv *CustomValidator) Validate(i any) error {
	if err := cv.V.Struct(i); err != nil {
		if fields, ok := FormatValidationErrors(err); ok {
			return &Error{Err: err, Fields: fields}
		}
		return err
	}
	return nil
}

// FormatValidationErrors extracts field->tag pairs from validator errors.
func FormatValidationErrors(err error) (map[string]string, bool) {
	var ve validator.ValidationErrors
	ok := errors.As(err, &ve)
	if !ok {
		return nil, false
	}
	out := make(map[string]string, len(ve))
	for _, fe := range ve {
		out[strings.ToLower(fe.Field())] = fe.Tag()
	}
	return out, true
}
