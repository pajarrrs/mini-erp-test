package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"mini-erp-test/backend/internal/handler"
	"mini-erp-test/backend/internal/repository"
)

func main() {
	databaseURL := getEnv("DATABASE_URL", "postgres://mini-erp-test:mini-erp-test@localhost:5432/mini-erp-test?sslmode=disable")
	port := getEnv("PORT", "8080")

	db, err := connectDB(databaseURL)
	if err != nil {
		log.Fatalf("cannot connect to database: %v", err)
	}
	defer db.Close()

	repo := repository.New(db)
	h := handler.New(repo)

	log.Printf("API listening on :%s", port)
	if err := http.ListenAndServe(":"+port, logRequests(h.Routes())); err != nil {
		log.Fatal(err)
	}
}

func connectDB(url string) (*pgxpool.Pool, error) {
	db, err := pgxpool.New(context.Background(), url)
	if err != nil {
		return nil, err
	}
	for attempt := 1; attempt <= 10; attempt++ {
		if err = db.Ping(context.Background()); err == nil {
			return db, nil
		}
		log.Printf("database not ready (attempt %d): %v", attempt, err)
		time.Sleep(2 * time.Second)
	}
	return nil, err
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
