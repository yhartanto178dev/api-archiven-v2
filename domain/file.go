package domain

import (
	"errors"
	"time"
)

type File struct {
	ID          string
	Name        string
	Size        int64
	ContentType string
	UploadDate  time.Time
}

// success
var Success = "success"

// Error
var ErrFileNotFound = errors.New("file not found")
