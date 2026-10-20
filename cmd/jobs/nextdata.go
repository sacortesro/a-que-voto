package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
)

var nextDataRe = regexp.MustCompile(`(?s)<script id="__NEXT_DATA__"[^>]*>(.*?)</script>`)

func parseNextData[T any](html []byte) (*T, error) {
	m := nextDataRe.FindSubmatch(html)
	if m == nil {
		return nil, errors.New("__NEXT_DATA__ not found")
	}
	var nd T
	if err := json.Unmarshal(m[1], &nd); err != nil {
		return nil, fmt.Errorf("decode __NEXT_DATA__: %w", err)
	}
	return &nd, nil
}
