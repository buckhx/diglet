// Package digletss is a diglet tile server
package burrow

import (
	"net/http"

	"github.com/gorilla/mux"
)

type App struct {
	Prefix  string
	Name    string
	Methods []Method
	Debug   bool
	Port    string
	Router  *mux.Router
}

func (app *App) Run() error {
	r := mux.NewRouter()
	r.StrictSlash(true)
	routes := &RouteHandler{Prefix: app.Prefix}
	routes.MountRoutes(app.Methods)
	routes.Subrouter(r)
	app.Router = r
	return app.start()
}

func NewApp(name string) *App {
	return &App{
		Name:   name,
		Prefix: "/",
		Debug:  false,
		Port:   "8080",
	}
}

func (app *App) start() (err error) {
	info("%s used Burrow!", app.Name)

	app.mountStatic()
	http.Handle("/", app.Router)

	info("Now serving on port :%s", app.Port)
	err = http.ListenAndServe(":"+app.Port, nil)
	check(err)
	return
}

func (app *App) mountStatic() {
	static := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
	app.Router.PathPrefix("/static/").Handler(static)
}
