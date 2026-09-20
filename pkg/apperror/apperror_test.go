package apperror

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConstructors_StatusCodes(t *testing.T) {
	assert.Equal(t, http.StatusUnauthorized, Unauthorized("x").Status)
	assert.Equal(t, "unauthorized", Unauthorized("x").Code)
	assert.Equal(t, http.StatusTooManyRequests, TooManyRequests("x").Status)
	assert.Equal(t, "rate_limited", TooManyRequests("x").Code)
	assert.Equal(t, http.StatusConflict, Conflict("x").Status)
	assert.Equal(t, http.StatusInternalServerError, Internal().Status)
}
