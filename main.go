package main

import (
    "fmt"
    "net/http"
)

func main() {
    mux := http.NewServeMux()

    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "Hello, Chirpy!")
    })

    server := &http.Server{
        Addr:    ":8080",
        Handler: mux,
    }

    fmt.Println("Server is running at http://localhost:8080")
    err := server.ListenAndServe()
    if err != nil {
        fmt.Println("Server error:", err)
    }
}
