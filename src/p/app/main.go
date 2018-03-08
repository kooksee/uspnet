package app

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)



// Run app run
func Run() error {
	router := httprouter.New()
	initUrls(router)

	return http.ListenAndServe(":9000", router)
}
