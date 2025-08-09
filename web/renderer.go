package web

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/flosch/pongo2/v6"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"playtime/web/localization"
	"strings"
)

type EmbedFileSystemLoader struct {
	fs fs.FS
}

func (l *EmbedFileSystemLoader) Abs(base, name string) string {
	return filepath.Join(filepath.Dir(base), name)
}

func (l *EmbedFileSystemLoader) Get(path string) (io.Reader, error) {
	data, err := fs.ReadFile(l.fs, path)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), nil
}

type pongo2Renderer struct {
	config *Configuration
	set    *pongo2.TemplateSet
}

func getAllFilenames(efs fs.FS) (files []string, err error) {
	if err := fs.WalkDir(efs, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
 
		files = append(files, path)

		return nil
	}); err != nil {
		return nil, err
	}
	log.Infof("%w", files)
	return files, nil
}

func newPongo2Renderer(config *Configuration) pongo2Renderer {
	loader := &EmbedFileSystemLoader{fs: config.TemplatesFS}
	set := pongo2.NewSet("echo", loader)
	localization.Init(config.AssetsFS)
	return pongo2Renderer{config: config, set: set}
}

func (r pongo2Renderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	var ctx pongo2.Context
	var ok bool
	if data != nil {
		ctx, ok = data.(pongo2.Context)
		if !ok {
			return errors.New("no pongo2.Context data was passed")
		}
	}

	var t *pongo2.Template
	var err error
	if r.config.TemplatesDebug {
		t, err = r.set.FromFile(r.resolveTemplateName(name))
	} else {
		t, err = r.set.FromCache(r.resolveTemplateName(name))
	}
	if err != nil {
		return err
	}

	ctx["split_description"] = func(s string) []string {
		return strings.Split(s, "\n")
	}

	_, ok = ctx["netplay_enabled"]
	if !ok {
		ctx["netplay_enabled"] = r.config.NetplayEnabled
	}

	//l10n
	lang := localization.Code(c)
	ctx["localization_list"] = localization.List()
	ctx["localization_lang"] = lang
	ctx["loc"] = func(s string, args ...any) string {
		return localization.Localize(lang, s, args)
	}

	return t.ExecuteWriter(ctx, w)
}

func (r pongo2Renderer) resolveTemplateName(n string) string {
	return fmt.Sprintf("%s%c%s.%s", r.config.TemplatesRoot, os.PathSeparator, n, r.config.TemplatesExtension)
}

func httpErrorHandler(e error, c echo.Context) {
	code := http.StatusInternalServerError
	if httpError, ok := e.(*echo.HTTPError); ok {
		code = httpError.Code
	}

	log.WithFields(log.Fields{
		"method": c.Request().Method,
		"uri":    c.Request().URL,
		"error":  e,
	}).Error("request error")

	if err := c.Render(code, "error", pongo2.Context{"error": e}); err != nil {
		log.Errorf("error page render error: %s", err)
	}
}
