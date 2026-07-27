package main

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"time"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
)

//go:embed web/**
var frontend embed.FS

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {

	mux := http.NewServeMux()

	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")

		time.Sleep(2 * time.Second)

		res := contracts.Response{
			Time: time.Now().Format("2006-01-02 15:04:05"),
		}

		json.NewEncoder(w).Encode(res)
	})

	webFS, err := fs.Sub(frontend, "web")
	if err != nil {
		log.Fatal(err)
	}

	mux.Handle("/", http.FileServer(http.FS(webFS)))

	handler := enableCORS(mux)

	log.Println("Server running on http://localhost:8080")

	log.Fatal(
		http.ListenAndServe(":8080", handler),
	)
}
