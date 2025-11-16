package server

import (
	"html/template"
	"net/http"

	f "github.com/fcjr/scroll-together/server/internal/frontend"
)

func (s *Server) index() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		t, err := template.ParseFS(f.EmbeddedFiles, "index.html")
		if err != nil {
			s.logger.Error("unable to parse index.html template", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/html")
		t.Execute(w, nil)
	}
}

func (s *Server) script() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		f, err := f.EmbeddedFiles.ReadFile("client.mjs")
		if err != nil {
			s.logger.Error("unable to read client.mjs from static fs", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/javascript")
		w.Write(f)
	}
}
