package validation_test

import (
	"errors"
	"testing"

	"ArifulProtik/TownHall/pkg/validation"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type sample struct {
	Email string `validate:"required,email"`
}

func TestValidateBadEmail(t *testing.T) {
	v := validation.New()
	err := v.Validate(&sample{Email: "not-an-email"})
	require.Error(t, err)

	ve := &validation.Error{}
	ok := errors.As(err, &ve)
	require.True(t, ok, "expected *validation.Error, got %T", err)
	assert.Equal(t, "email", ve.Fields["email"])
}

func TestValidateGoodEmail(t *testing.T) {
	v := validation.New()
	assert.NoError(t, v.Validate(&sample{Email: "joe@example.com"}))
}

func TestValidateSignupShape(t *testing.T) {
	type signup struct {
		Name     string `validate:"required,min=2,max=100"`
		Email    string `validate:"required,email,max=255"`
		Password string `validate:"required,min=8,max=72"`
	}

	tests := []struct {
		name    string
		input   signup
		wantErr bool
		field   string
	}{
		{"valid", signup{"Joe", "joe@example.com", "password123"}, false, ""},
		{"bad email", signup{"Joe", "not-an-email", "password123"}, true, "email"},
		{"short password", signup{"Joe", "joe@example.com", "short"}, true, "password"},
		{"missing name", signup{"", "joe@example.com", "password123"}, true, "name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validation.New()
			err := v.Validate(&tt.input)
			if !tt.wantErr {
				assert.NoError(t, err)
				return
			}
			require.Error(t, err)
			ve := &validation.Error{}
			ok := errors.As(err, &ve)
			require.True(t, ok, "expected *validation.Error, got %T", err)
			assert.Contains(t, ve.Fields, tt.field)
		})
	}
}

func TestAlphanumUnderscoreValidation(t *testing.T) {
	v := validation.New()
	type TestPayload struct {
		Username string `validate:"alphanum_underscore"`
	}
	require.NoError(t, v.Validate(&TestPayload{Username: "valid_name123"}))
	require.Error(t, v.Validate(&TestPayload{Username: "invalid-name!"}))
}
