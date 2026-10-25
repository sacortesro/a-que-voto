package auth

import (
	"errors"
	"net/http"
	"strings"
)

// GetAPIKey extracts the API token from the Authorization header.
//
//	Authorization: Bearer <token>
func GetAPIKey(headers http.Header) (string, error) {
	val := headers.Get("Authorization")
	if val == "" {
		return "", errors.New("no authentication info found")
	}

	vals := strings.Fields(val)
	if len(vals) != 2 || !strings.EqualFold(vals[0], "Bearer") {
		return "", errors.New("invalid authentication header")
	}

	return vals[1], nil
}
