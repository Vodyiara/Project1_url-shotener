package storage

import "errors"

var (
	ErrUrlExists   = errors.New("url exists")
	ErrURLNotFound = errors.New("url not found")
)
