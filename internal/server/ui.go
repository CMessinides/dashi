package server

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"maps"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

type ui struct {
	Assets   fs.FS
	Renderer echo.Renderer
	config   uiConfig
}

type uiConfig struct {
	Dev           bool
	AssetPath     string
	TemplateFuncs template.FuncMap
}

func (c uiConfig) WithTemplateFuncs(funcs template.FuncMap) uiConfig {
	var merged template.FuncMap
	if c.TemplateFuncs == nil {
		merged = template.FuncMap{}
	} else {
		merged = maps.Clone(c.TemplateFuncs)
	}

	maps.Copy(merged, funcs)

	return uiConfig{
		Dev:           c.Dev,
		AssetPath:     c.AssetPath,
		TemplateFuncs: merged,
	}
}

func newUI(conf uiConfig) *ui {
	a := newAssetsFS(conf)

	conf = conf.WithTemplateFuncs(a.TemplateFuncs())
	t := newTemplateRenderer(conf)

	return &ui{
		Assets:   a,
		Renderer: t,
		config:   conf,
	}
}

func (u *ui) ConfigureServer(e *echo.Echo) {
	e.StaticFS(u.config.AssetPath, u.Assets)
	e.Renderer = u.Renderer
}

//go:embed all:public
var assets embed.FS

type assetsFS struct {
	publicPath string
	fs         fs.FS
}

func newAssetsFS(conf uiConfig) *assetsFS {
	a := &assetsFS{
		publicPath: conf.AssetPath,
	}

	if conf.Dev {
		a.fs = os.DirFS("internal/server/public")
	} else {
		a.fs = echo.MustSubFS(assets, "public")
	}

	return a
}

func (a *assetsFS) Open(name string) (fs.File, error) {
	return a.fs.Open(name)
}

func (a *assetsFS) FileExists(name string) (bool, error) {
	file, err := a.Open(name)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		} else {
			return false, err
		}
	}

	info, err := file.Stat()
	if err != nil {
		return false, err
	}

	return !info.IsDir(), nil
}

func (a *assetsFS) TemplateFuncs() template.FuncMap {
	asset := func(filename string) string {
		return a.publicPath + "/" + filename
	}

	stylesheet := func(filename string) template.HTML {
		return template.HTML(
			fmt.Sprintf(
				`<link rel="stylesheet" href="%s/dist/%s">`,
				template.HTMLEscapeString(a.publicPath),
				template.HTMLEscapeString(filename),
			),
		)
	}

	script := func(filename string) template.HTML {
		return template.HTML(
			fmt.Sprintf(
				`<script defer type="module" src="%s/dist/%s"></script>`,
				template.HTMLEscapeString(a.publicPath),
				template.HTMLEscapeString(filename),
			),
		)
	}

	return template.FuncMap{
		"asset":      asset,
		"stylesheet": stylesheet,
		"script":     script,
		"assetIfExists": func(filename string) string {
			if exists, _ := a.FileExists("dist/" + filename); !exists {
				return ""
			}

			return asset(filename)
		},
		"stylesheetIfExists": func(filename string) template.HTML {
			if exists, _ := a.FileExists("dist/" + filename); !exists {
				return ""
			}

			return stylesheet(filename)
		},
		"scriptIfExists": func(filename string) template.HTML {
			if exists, _ := a.FileExists(filename); !exists {
				return ""
			}

			return script(filename)
		},
		"icon": func(id string) template.HTML {
			return template.HTML(
				fmt.Sprintf(
					`<svg class="icon" height="16" width="16"><use xlink:href="%s/icons.svg#%s"></use></svg>`,
					a.publicPath,
					id,
				),
			)
		},
	}
}

type devTemplate struct {
	funcs template.FuncMap
}

func (d *devTemplate) ExecuteTemplate(w io.Writer, name string, data any) error {
	t := template.New("default.html").Funcs(d.funcs)
	t, err := t.ParseFS(os.DirFS("internal/server/views/layouts"), "*.html")
	if err != nil {
		return err
	}

	t, err = t.ParseFS(os.DirFS("internal/server/views/partials"), "*.html")
	if err != nil {
		return err
	}

	p := parseTemplateName(name)

	t, err = t.ParseFiles("internal/server/views/" + p.File)
	if err != nil {
		return err
	}

	return t.ExecuteTemplate(w, p.Template, data)
}

//go:embed views/*.html
var views embed.FS

//go:embed views/layouts/*.html
var layouts embed.FS

//go:embed views/partials/*.html
var partials embed.FS

type embeddedTemplate struct {
	shared *template.Template
}

func (e *embeddedTemplate) ExecuteTemplate(w io.Writer, name string, data any) error {
	shared, err := e.shared.Clone()
	if err != nil {
		return err
	}

	p := parseTemplateName(name)

	t, err := shared.ParseFS(views, "views/"+p.File)
	if err != nil {
		return err
	}

	return t.ExecuteTemplate(w, p.Template, data)
}

func newTemplateRenderer(conf uiConfig) *echo.TemplateRenderer {
	f := template.FuncMap{
		"formatISOTimestamp": func(t time.Time) string {
			return t.Format(time.RFC3339)
		},
		"formatRelativeTime": func(t time.Time) string {
			return time.Since(t).String()
		},
	}

	maps.Copy(f, conf.TemplateFuncs)

	if !conf.Dev {
		shared := template.New("default.html").Funcs(f)
		shared = template.Must(shared.ParseFS(layouts, "views/layouts/*.html"))
		shared = template.Must(shared.ParseFS(partials, "views/partials/*.html"))
		return &echo.TemplateRenderer{
			Template: &embeddedTemplate{
				shared: shared,
			},
		}
	} else {
		return &echo.TemplateRenderer{
			Template: &devTemplate{funcs: f},
		}
	}
}

type templateNameParts struct {
	File     string
	Template string
}

func parseTemplateName(name string) templateNameParts {
	p := templateNameParts{
		File:     name,
		Template: name,
	}

	if strings.Contains(name, "#") {
		parts := strings.SplitN(name, "#", 2)
		p.File = parts[0]
		p.Template = parts[1]
	}

	return p
}
