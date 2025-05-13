package server

import (
	"bytes"
	"io"
	"net/http"
)

type Config struct {
	Dev bool
}

type Server struct {
	Assets   *AssetHandler
	Renderer Renderer
	mux      *http.ServeMux
}

func NewServer(conf Config) *Server {
	s := &Server{
		mux: http.NewServeMux(),
		Assets: defaultAssetHandler(AssetConfig{
			PublicPath: "/assets",
			Dev:        conf.Dev,
		}),
	}

	s.Renderer = defaultRenderer(RendererConfig{
		Dev:   conf.Dev,
		Funcs: s.Assets.TemplateFuncs(),
	})

	s.mux.HandleFunc("GET /", s.dashboard)

	s.mux.Handle("GET "+s.Assets.PublicPath+"/", http.FileServerFS(s.Assets.FS))

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.mux.ServeHTTP(w, req)
}

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	buf := new(bytes.Buffer)
	v := NewView("dashboard", map[string]any{
		"View": "dashboard",
	})

	err := s.render(buf, v)

	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
	} else {
		w.Header().Add("Content-Type", "text/html;utf8")
		w.WriteHeader(200)
		w.Write(buf.Bytes())
	}
}

func (s *Server) render(w io.Writer, view *View) error {
	return s.Renderer.Render(w, view)
}
