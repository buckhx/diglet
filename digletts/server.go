// Package digletss is a diglet tile server
package digletts

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Server struct {
	DataDir, Port string
	Router        *mux.Router
}

func (s *Server) Start() (err error) {
	info("Starting server...")

	s.mountStatic()
	http.Handle("/", s.Router)

	info("Now serving tiles from %s on port %s", s.DataDir, s.Port)
	err = http.ListenAndServe(s.Port, nil)
	check(err)
	return
}

func (s *Server) mountStatic() {
	static := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
	s.Router.PathPrefix("/static/").Handler(static)
}
