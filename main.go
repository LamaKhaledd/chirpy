package main

import (
    "net/http"
)

func readinessHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}

func main() {
    mux := http.NewServeMux()

    mux.HandleFunc("/healthz", readinessHandler)

    fs := http.FileServer(http.Dir("app"))
    mux.Handle("/app/", http.StripPrefix("/app", fs))

    http.ListenAndServe(":8080", mux)
}
