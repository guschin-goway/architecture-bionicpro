package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	_ "github.com/ClickHouse/clickhouse-go"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/justinas/alice"
)

type ReportRow struct {
	ReportDate        string  `json:"report_date"`
	ProsthesisID      int     `json:"prosthesis_id"`
	MovementsCount    int     `json:"movements_count"`
	AvgReactionTimeMs float64 `json:"avg_reaction_time_ms"`
	BatteryAvgLevel   float64 `json:"battery_avg_level"`
	ErrorsCount       int     `json:"errors_count"`
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// OPTIONS-запросы пропускаем для CORS preflight
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Просто сохраняем заголовок в контексте, чтобы handler мог его использовать
		ctx := context.WithValue(r.Context(), "Authorization", authHeader)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if origin == "http://localhost:3000" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set(
				"Access-Control-Allow-Headers",
				"Content-Type, X-User-ID, Authorization", // <-- добавили Authorization
			)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// Preflight OPTIONS-запрос
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	db, err := sql.Open("clickhouse", "clickhouse://clickhouse:9000")
	if err != nil {
		log.Fatal("Failed to connect to ClickHouse:", err)
	}

	reportsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(`
			SELECT
				report_date,
				prosthesis_id,
				movements_count,
				avg_reaction_time_ms,
				battery_avg_level,
				errors_count
			FROM report_user_daily
			ORDER BY report_date DESC
			LIMIT 10
		`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var result []ReportRow
		for rows.Next() {
			var r ReportRow
			if err := rows.Scan(
				&r.ReportDate,
				&r.ProsthesisID,
				&r.MovementsCount,
				&r.AvgReactionTimeMs,
				&r.BatteryAvgLevel,
				&r.ErrorsCount,
			); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			result = append(result, r)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	// CORS middleware (должен быть ПЕРВЫМ)

	// Цепочка middleware
	chain := alice.New(
		corsMiddleware, // <-- Ставим первым
		authMiddleware, // аутентификация
		middleware.RequestID,
		middleware.RealIP,
		middleware.Logger,
		middleware.Recoverer,
	).Then(reportsHandler)

	http.Handle("/reports", chain)

	log.Println("Reports API started on :8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
