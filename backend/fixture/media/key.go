// Package media provides storage-oriented fixture helpers for media workflows.
package media

import (
	"path"
	"strings"
)

const (
	// PrefixUsers is the root prefix for user-owned media objects.
	PrefixUsers = "users"
	// PrefixUploads is the root prefix for transient upload objects.
	PrefixUploads = "uploads"
)

// ObjectKey joins path segments into a normalized object key.
func ObjectKey(parts ...string) string {
	cleaned := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.Trim(part, "/")
		if part != "" {
			cleaned = append(cleaned, part)
		}
	}
	return path.Join(cleaned...)
}

// UserObjectKey builds users/{userID}/{filename}.
func UserObjectKey(userID, filename string) string {
	return ObjectKey(PrefixUsers, userID, filename)
}

// UploadObjectKey builds uploads/{uploadID}/{filename}.
func UploadObjectKey(uploadID, filename string) string {
	return ObjectKey(PrefixUploads, uploadID, filename)
}

// AudioObjectKey builds users/{userID}/audio/{recordingID}.wav.
func AudioObjectKey(userID, recordingID string) string {
	return ObjectKey(PrefixUsers, userID, "audio", recordingID+".wav")
}
