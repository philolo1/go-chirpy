package main

import (
	"fmt"
	http "net/http"

	"github.com/go-chi/chi/v5"
)

type ApiConfig struct {
	fileserverHits int
}

func (cfg *ApiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits += 1
		next.ServeHTTP(w, r)
	})
}

func (cfg *ApiConfig) metrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`
<html>

<body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
</body>

</html>
`, cfg.fileserverHits)))
}

func (cfg *ApiConfig) reset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	cfg.fileserverHits = 0
	w.Write([]byte("OK"))
}

func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Request: %v\n", r.URL)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func apiRouter(apiConfig *ApiConfig) http.Handler {
	r := chi.NewRouter()

	r.HandleFunc("/reset", apiConfig.reset)

	r.Get("/healthz", healthz)

	return r
}

func adminRouter(apiConfig *ApiConfig) http.Handler {
	r := chi.NewRouter()

	r.Get("/metrics", apiConfig.metrics)

	return r
}

func main() {
	apiConfig := ApiConfig{}

	r := chi.NewRouter()
	fsHandler := http.StripPrefix("/app", apiConfig.middlewareMetricsInc(http.FileServer(http.Dir("."))))

	r.Handle("/app", fsHandler)
	r.Handle("/app/*", fsHandler)

	r.Mount("/api", apiRouter(&apiConfig))
	r.Mount("/admin", adminRouter(&apiConfig))

	corsMux := middlewareCors(r)

	server := http.Server{
		Handler: corsMux,
		Addr:    ":8080",
	}

	fmt.Println("Start server at localhost:8080")
	server.ListenAndServe()

}
