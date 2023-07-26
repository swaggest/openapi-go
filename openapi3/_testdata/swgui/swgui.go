package main

import (
	"log"
	"net/http"
	"os"

	v3 "github.com/swaggest/swgui/v5"
)

func main() {
	h := v3.NewHandler("Foo", "/openapi.json", "/")
	hh := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/openapi.json" {
			o, err := os.ReadFile("../openapi.json")
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
	_ = http.ListenAndServe("localhost:8082", hh)
}
