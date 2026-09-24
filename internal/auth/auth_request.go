package auth

import (
	"errors"
	"net/http"
	"strings"
)

var errNoBearer = errors.New("missing bearer token")

// ViewerID returns the id in the request's bearer token, or "" when the
// request is anonymous or the token is bad. For pages anyone may read.
func ViewerID(r *http.Request, secret string) string {
	uid, err := bearerUserID(r, secret)
	if err != nil {
		return ""
	}
	return uid
}

// bearerUserID checks a request's Authorization header and returns the id
// inside a valid access token.
func bearerUserID(r *http.Request, secret string) (string, error) {
	header := r.Header.Get("Authorization")
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") || parts[1] == "" {
		return "", errNoBearer
	}
	return VerifyAccessToken(secret, parts[1])
}
