package domain

import "errors"

var (
	ErrInputisEmpty = errors.New("input is empty")
	ErrNotFound     = errors.New("key is not found")
	ErrInvalidId    = errors.New("invalid ID")
)
