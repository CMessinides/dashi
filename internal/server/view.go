package server

import (
	"embed"
	"html/template"
	"io"
	"os"
)

type View struct {
	Name  string
	Block string
	Data  any
}

func NewView(name string, data any) *View {
	return &View{
		Name: name,
		Data: data,
	}
}

func (v *View) WithBlock(block string) *View {
	return &View{
		Name:  v.Name,
		Block: block,
		Data:  v.Data,
	}
}

type Renderer interface {
	Render(w io.Writer, view *View) error
}

type RendererConfig struct {
	Dev   bool
	Funcs template.FuncMap
}

type devTemplate struct {
	funcs template.FuncMap
}

func (d *devTemplate) Render(w io.Writer, view *View) error {
	t := template.New(view.Name).Funcs(d.funcs)
	t, err := t.ParseFS(os.DirFS("internal/server/views/layouts"), "*.html")
	if err != nil {
		return err
	}

	t, err = t.ParseFS(os.DirFS("internal/server/views/partials"), "*.html")
	if err != nil {
		return err
	}

	t, err = t.ParseFiles("internal/server/views/" + view.Name + ".html")
	if err != nil {
		return err
	}

	name := view.Block
	if name == "" {
		name = view.Name + ".html"
	}

	return t.ExecuteTemplate(w, name, view.Data)
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

func (e *embeddedTemplate) Render(w io.Writer, view *View) error {
	shared, err := e.shared.Clone()
	if err != nil {
		return err
	}

	t, err := shared.ParseFS(views, "views/"+view.Name+".html")
	if err != nil {
		return err
	}

	name := view.Block
	if name == "" {
		name = view.Name + ".html"
	}

	return t.ExecuteTemplate(w, name, view.Data)
}

func defaultRenderer(conf RendererConfig) Renderer {
	if !conf.Dev {
		shared := template.New("default.html").Funcs(conf.Funcs)
		shared = template.Must(shared.ParseFS(layouts, "views/layouts/*.html"))
		shared = template.Must(shared.ParseFS(partials, "views/partials/*.html"))
		return &embeddedTemplate{
			shared: shared,
		}
	} else {
		return &devTemplate{
			funcs: conf.Funcs,
		}
	}
}
