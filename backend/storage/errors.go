package storage

import "errors"

// ErrNotFound indicates the requested object or version does not exist.
var ErrNotFound = errors.New("storage: object not found")

// ErrVersioningDisabled indicates a versioned operation was requested but versioning is off.
var ErrVersioningDisabled = errors.New("storage: versioning disabled")
