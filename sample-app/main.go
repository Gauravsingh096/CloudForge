package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		name := os.Getenv("APP_NAME")
		if name == "" {
			name = "sample-app"
		}
		fmt.Fprintf(w, "Hello from CloudForge! App: %s\n", name)
	})

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("sample-app listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
