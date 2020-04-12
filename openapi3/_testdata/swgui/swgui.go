package main

import (
	"io/ioutil"
	"log"
	"net/http"

	v3 "github.com/swaggest/swgui/v3"
)

func main() {
	h := v3.NewHandler("Foo", "/openapi.json", "/")
	hh := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/openapi.json" {
			o, err := ioutil.ReadFile("openapi3/_testdata/openapi.json")
			if err != nil {
				http.Error(rw, err.Error(), 500)
				return
			}
			rw.Header().Set("Content-Type", "application/json")
			rw.Write(o)
			return
		}

		h.ServeHTTP(rw, r)
	})
	log.Println("Starting Swagger UI server at http://localhost:8082/")
	http.ListenAndServe(":8082", hh)
}
