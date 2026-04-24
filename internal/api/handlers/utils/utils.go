package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func Decode[T any](r *http.Request) (T, error) {
	var req T

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		return req, fmt.Errorf("failed to decode json: %w", err)
	}

	return req, nil
}
