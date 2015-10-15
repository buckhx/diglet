package digletts

import (
	"net/http"
)

func check(w http.ResponseWriter, err error) (caught bool) {
	caught = false
	if err != nil {
		caught = true
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	return
}

func checks(errs ...error) {
	for _, err := range errs {
		if err != nil {
			panic(err)
		}
	}
}
