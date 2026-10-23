package utils

import (
	"database/sql"
	"time"
)

const DateLayout = "2006-01-02"

func ParseDate(s string) (time.Time, error) {
	return time.Parse(DateLayout, s)
}

func ParseOptionalDate(s *string) (*time.Time, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	t, err := ParseDate(*s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func FormatDate(t time.Time) string {
	return t.Format(DateLayout)
}

func FormatTimestamp(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func FormatNullDate(t sql.NullTime) string {
	if !t.Valid {
		return ""
	}
	return FormatDate(t.Time)
}
