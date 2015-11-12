// Package digletss is a diglet tile server
package ioserver

import (
	"net/http"

	"github.com/gorilla/mux"
)

type App struct {
	Prefix  string
	Methods []Method
	Server  *Server
}

func (app *App) Run() error {
	r := mux.NewRouter()
	r.StrictSlash(true)
	routes := &RouteHandler{Prefix: app.Prefix}
	routes.MountRoutes(app.Methods)
	routes.Subrouter(r)
	app.Server.Router = r
	return app.Server.Start()
}

func NewApp(dir string, port string) *App {
	return &App{
		Prefix: "/",
		Server: &Server{
			DataDir: dir,
			Port:    ":" + port,
		},
	}
}

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
