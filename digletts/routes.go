package digletts

import (
	"net/http"

	"github.com/gorilla/mux"
)

type HandleFunc func(w http.ResponseWriter, r *http.Request)

var routes = []struct {
	Pattern string
	Handler HandleFunc
}{
	{"/tileset/{z}/{x}/{y}", TileHandler},
	{"/tileset/metadata", MetadataHandler},
}

func BuildRouter() *mux.Router {
	r := mux.NewRouter()
	for _, route := range routes {
		r.HandleFunc(route.Pattern, route.Handler)
	}
	return r
}
