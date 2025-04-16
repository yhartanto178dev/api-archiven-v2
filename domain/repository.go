package domain

import "io"

type FileRepository interface {
	Save(file *File, content io.Reader) error
	FindByID(id string) (*File, io.ReadCloser, error)
	FindAll(skip, limit int64) ([]*File, error)
	Count() (int64, error)
}
