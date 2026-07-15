package storage

import (
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

type localFileStore struct {
	uploadsPath string
}

func newLocalFileStore(uploadsPath string) *localFileStore {
	return &localFileStore{uploadsPath: uploadsPath}
}

func (l *localFileStore) fullPath(id, extension string) (string, error) {
	uploadPath, err := GetUploadPath(id)
	if err != nil {
		return "", err
	}
	fileName := id
	if extension != "" {
		fileName += "." + extension
	}
	return filepath.Join(l.uploadsPath, filepath.FromSlash(uploadPath), fileName), nil
}

func (l *localFileStore) Save(file *multipart.FileHeader, id, extension string) error {
	uploadPath, err := GetUploadPath(id)
	if err != nil {
		return err
	}

	dirPath := filepath.Join(l.uploadsPath, filepath.FromSlash(uploadPath))
	if err := os.MkdirAll(dirPath, 0777); err != nil {
		return err
	}

	fileName := id
	if extension != "" {
		fileName += "." + extension
	}

	src, err := file.Open()
	if err != nil {
		return err
	}
	defer func() { _ = src.Close() }()

	dst, err := os.Create(filepath.Join(dirPath, fileName))
	if err != nil {
		return err
	}
	defer func() { _ = dst.Close() }()

	_, err = io.Copy(dst, src)
	return err
}

func (l *localFileStore) Delete(id, extension string) error {
	p, err := l.fullPath(id, extension)
	if err != nil {
		return err
	}
	return os.Remove(p)
}

func (l *localFileStore) Open(id, extension string) (io.ReadCloser, int64, error) {
	p, err := l.fullPath(id, extension)
	if err != nil {
		return nil, 0, err
	}

	f, err := os.Open(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, 0, os.ErrNotExist
		}
		return nil, 0, err
	}

	stat, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, 0, err
	}

	return f, stat.Size(), nil
}

func (l *localFileStore) SaveMeta(kind, id string, data []byte) error { return nil }
func (l *localFileStore) DeleteMeta(kind, id string) error              { return nil }
func (l *localFileStore) ListMeta(kind string) ([][]byte, error)        { return nil, nil }

func (l *localFileStore) Head(id, extension string) (int64, error) {
	p, err := l.fullPath(id, extension)
	if err != nil {
		return 0, err
	}

	stat, err := os.Stat(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, os.ErrNotExist
		}
		return 0, err
	}

	return stat.Size(), nil
}
