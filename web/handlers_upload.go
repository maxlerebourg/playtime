package web

import (
	"errors"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

func parseUploadParam(c echo.Context) (id, extension string) {
	name := c.Param("*")
	parts := strings.Split(name, "/")
	last := parts[len(parts)-1]
	if dot := strings.LastIndex(last, "."); dot >= 0 {
		return last[:dot], last[dot+1:]
	}
	return last, ""
}

func (s *Server) uploadsGet(c echo.Context) error {
	id, ext := parseUploadParam(c)

	reader, size, err := s.storage.OpenUpload(id, ext)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return echo.ErrNotFound
		}
		return err
	}
	defer func() { _ = reader.Close() }()

	contentType := mime.TypeByExtension("." + ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	c.Response().Header().Set("Content-Length", strconv.FormatInt(size, 10))
	return c.Stream(http.StatusOK, contentType, reader)
}

func (s *Server) uploadsHead(c echo.Context) error {
	id, ext := parseUploadParam(c)

	size, err := s.storage.HeadUpload(id, ext)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return echo.ErrNotFound
		}
		return err
	}

	c.Response().Header().Add("Content-Length", fmt.Sprintf("%d", size))
	return c.String(http.StatusOK, "")
}

func (s *Server) assetsHead(c echo.Context) error {
	name := filepath.ToSlash(filepath.Clean(strings.TrimPrefix(c.Param("*"), "/")))

	assetsRoot, err := filepath.Abs(s.config.AssetsRoot)
	if err != nil {
		return err
	}

	assetPath, err := filepath.Abs(path.Join(assetsRoot, name))
	if err != nil {
		return err
	}

	if !startsWith(assetPath, assetsRoot) {
		return errors.New("not in directory")
	}

	stat, err := os.Stat(assetPath)
	if err != nil {
		return err
	}

	c.Response().Header().Add("Content-Length", fmt.Sprintf("%d", stat.Size()))

	return c.String(http.StatusOK, "")
}
