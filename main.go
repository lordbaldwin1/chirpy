package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/lordbaldwin1/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	queries        *database.Queries
	platform       string
	jwtSecret      string
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("fatal: %s", err)
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}
	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFORM must be set")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET must be set")
	}

	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("fatal: %s", err)
	}
	dbQueries := database.New(dbConn)

	const filePathRoot = "."
	const port = "8080"

	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		queries:        dbQueries,
		platform:       platform,
		jwtSecret:      jwtSecret,
	}

	mux := http.NewServeMux()

	fileServerHandler := http.StripPrefix("/app", http.FileServer(http.Dir(filePathRoot)))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(fileServerHandler))

	mux.HandleFunc("GET /api/healthz", healthzHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	mux.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerChirpsCreate)
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerChirpsGet)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerChirpsGetByID)
	mux.HandleFunc("POST /api/login", apiCfg.handlerUsersLogin)
	mux.HandleFunc("POST /api/refresh", apiCfg.handleRefreshToken)
	mux.HandleFunc("POST /api/revoke", apiCfg.handlerRevoke)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port %s\n", filePathRoot, port)
	log.Fatal(server.ListenAndServe())
}
