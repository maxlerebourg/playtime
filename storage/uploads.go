package storage

import (
	"errors"
	"io"
	"mime/multipart"
	"regexp"
	"strings"
)

const (
	FileExtensionSaveState  = "sav"
	FileExtensionScreenshot = "png"
)

func (s *Storage) SaveUploadedFile(file *multipart.FileHeader, id, extension string) error {
	return s.fileStore.Save(file, id, extension)
}

func (s *Storage) removeUploadedFile(id, extension string) error {
	return s.fileStore.Delete(id, extension)
}

func (s *Storage) OpenUpload(id, extension string) (io.ReadCloser, int64, error) {
	return s.fileStore.Open(id, extension)
}

func (s *Storage) HeadUpload(id, extension string) (int64, error) {
	return s.fileStore.Head(id, extension)
}

// GetUploadPath returns the slash-separated directory path for a given ID,
// used both for constructing file paths and URL segments.
func GetUploadPath(id string) (string, error) {
	if len(id) == 0 {
		return "", errors.New("id is empty")
	}

	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		return "", err
	}

	filtered := reg.ReplaceAllString(id, "")

	var parts []string
	for i := 0; i < len(filtered); i += 2 {
		if i+2 <= len(filtered) {
			parts = append(parts, string([]rune(filtered)[i:i+2]))
		} else {
			parts = append(parts, string([]rune(filtered)[i:i+1]))
		}
	}

	return strings.Join(parts, "/"), nil
}
