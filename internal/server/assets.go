package server

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"os"
)

//go:embed all:public
var assets embed.FS

type AssetHandler struct {
	FS         fs.FS
	PublicPath string
}

func (a *AssetHandler) Open(name string) (fs.File, error) {
	return a.FS.Open(name)
}

func (a *AssetHandler) fileExists(name string) (bool, error) {
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

func (a *AssetHandler) TemplateFuncs() template.FuncMap {
	asset := func(filename string) string {
		return a.PublicPath + "/" + filename
	}

	stylesheet := func(filename string) template.HTML {
		return template.HTML(
			fmt.Sprintf(
				`<link rel="stylesheet" href="%s/dist/%s">`,
				template.HTMLEscapeString(a.PublicPath),
				template.HTMLEscapeString(filename),
			),
		)
	}

	script := func(filename string) template.HTML {
		return template.HTML(
			fmt.Sprintf(
				`<script defer type="module" src="%s/dist/%s"></script>`,
				template.HTMLEscapeString(a.PublicPath),
				template.HTMLEscapeString(filename),
			),
		)
	}

	return template.FuncMap{
		"asset":      asset,
		"stylesheet": stylesheet,
		"script":     script,
		"assetIfExists": func(filename string) string {
			if exists, _ := a.fileExists("dist/" + filename); !exists {
				return ""
			}

			return asset(filename)
		},
		"stylesheetIfExists": func(filename string) template.HTML {
			if exists, _ := a.fileExists("dist/" + filename); !exists {
				return ""
			}

			return stylesheet(filename)
		},
		"scriptIfExists": func(filename string) template.HTML {
			if exists, _ := a.fileExists(filename); !exists {
				return ""
			}

			return script(filename)
		},
		"icon": func(id string) template.HTML {
			return template.HTML(
				fmt.Sprintf(
					`<svg class="icon" height="16" width="16"><use xlink:href="%s/icons.svg#%s"></use></svg>`,
					a.PublicPath,
					id,
				),
			)
		},
	}
}

type AssetConfig struct {
	Dev        bool
	PublicPath string
}

func defaultAssetHandler(conf AssetConfig) *AssetHandler {
	a := &AssetHandler{
		PublicPath: conf.PublicPath,
	}

	if conf.Dev {
		a.FS = mustSubFS(assets, "public")
	} else {
		a.FS = os.DirFS("internal/server/public")
	}

	return a
}

func mustSubFS(fsys fs.FS, dir string) fs.FS {
	s, err := fs.Sub(fsys, dir)
	if err != nil {
		panic(err)
	}

	return s
}
