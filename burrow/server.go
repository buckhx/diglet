// Package digletss is a diglet tile server
package burrow

import (
	"net/http"

	"github.com/gorilla/mux"
)

type App struct {
	Prefix         string
	Name           string
	Methods        []Method
	Debug          bool
	Port           string
	Router         *mux.Router
	TLSCertificate *string
	TLSPrivateKey  *string
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

func (app *App) RunTLS(cert, key *string) error {
	app.TLSCertificate = cert
	app.TLSPrivateKey = key
	return app.Run()
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
	if app.hasCerts() {
		err = app.serveTLS()
	} else {
		err = app.serve()
	}
	check(err)
	return
}

func (app *App) serve() (err error) {
	info("Now serving unencrypted HTTP traffic on port :%s", app.Port)
	return http.ListenAndServe(app.GetPort(), nil)
}

func (app *App) serveTLS() (err error) {
	port := app.GetPort()
	cert := *app.TLSCertificate
	key := *app.TLSPrivateKey
	info("Now serving encrypted TLS traffic on port :%s", app.Port)
	return http.ListenAndServeTLS(port, cert, key, nil)
}

func (app *App) mountStatic() {
	static := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
	app.Router.PathPrefix("/static/").Handler(static)
}

func (app *App) hasCerts() bool {
	return (app.TLSCertificate != nil) && (app.TLSPrivateKey != nil)
}

func (app *App) GetPort() string {
	return ":" + app.Port
}
