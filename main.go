package main

import (
	"fmt"
	"html/template"
	"net/http"
	"sync/atomic"
	"encoding/json"
	"log"
	"strings"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	count := cfg.fileserverHits.Load()

	tmpl, err := template.ParseFiles("app/admin_metrics.html")
	if err != nil {
		http.Error(w, "Failed to load metrics page", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	data := struct {
		Count int32
	}{
		Count: count,
	}

	tmpl.Execute(w, data)
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits counter reset"))
}

func handlerHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

type chirpRequest struct {
	Body string `json:"body"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type validResponse struct {
	Valid bool `json:"valid"`
}
func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	var chirp chirpRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&chirp)
	if err != nil {
		log.Printf("Error decoding chirp: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "Something went wrong"})
		return
	}

	if len(chirp.Body) > 140 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "Chirp is too long"})
		return
	}

	// Profanity filter: replace profane words with ****
	profaneWords := []string{"kerfuffle", "sharbert", "fornax"}
	words := strings.Split(chirp.Body, " ")

	for i, word := range words {
		for _, bad := range profaneWords {
			if strings.EqualFold(word, bad) {
				words[i] = "****"
			}
		}
	}

	cleaned := strings.Join(words, " ")

	// Return cleaned body
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"cleaned_body": cleaned,
	})
}


func main() {
	apiCfg := &apiConfig{}

	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("app"))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", fileServer)))

	mux.Handle("GET /healthz", http.HandlerFunc(handlerHealthz))

	mux.Handle("GET /admin/metrics", http.HandlerFunc(apiCfg.handlerMetrics))
	mux.Handle("POST /admin/reset", http.HandlerFunc(apiCfg.handlerReset))
	mux.Handle("POST /api/validate_chirp", http.HandlerFunc(handlerValidateChirp))


	fmt.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", mux)
}
