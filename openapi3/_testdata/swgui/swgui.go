package main

import (
	"log"
	"net/http"

	swgui "github.com/swaggest/swgui/v5"
)

func main() {
	urlToSchema := "/openapi.json"
	filePathToSchema := "../openapi.json"

	swh := swgui.NewHandler("Foo", urlToSchema, "/")
	hh := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.URL.Path == urlToSchema {
			http.ServeFile(rw, r, filePathToSchema)
		}

		swh.ServeHTTP(rw, r)
	})

	log.Println("Starting Swagger UI server at http://localhost:8082/")
	_ = http.ListenAndServe("localhost:8082", hh)
}
