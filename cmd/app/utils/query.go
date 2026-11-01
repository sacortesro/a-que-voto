package utils

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

func ParseQueryPositiveInt(r *http.Request, key string, def int, max int) (int, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return def, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 || n > max {
		return 0, fmt.Errorf("%s must be an integer between 1 and %d", key, max)
	}
	return n, nil
}

func ParseQueryDate(r *http.Request, key string, def time.Time) (time.Time, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return def, nil
	}
	d, err := ParseDate(raw)
	if err != nil {
		return time.Time{}, fmt.Errorf("%s must be a date of format YYYY-MM-DD", key)
	}
	return d, nil
}
