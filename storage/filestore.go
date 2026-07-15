package storage

import (
	"io"
	"mime/multipart"
)

type FileStore interface {
	Save(file *multipart.FileHeader, id, extension string) error
	Delete(id, extension string) error
	Open(id, extension string) (io.ReadCloser, int64, error)
	Head(id, extension string) (int64, error)
	SaveMeta(kind, id string, data []byte) error
	DeleteMeta(kind, id string) error
	ListMeta(kind string) ([][]byte, error)
}
