package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	_ "github.com/ClickHouse/clickhouse-go"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"github.com/justinas/alice"
)

type ReportRow struct {
	UserID      string    `json:"user_id"`
	SensorValue string    `json:"sensor_value"`
	Ts          time.Time `json:"ts"`
	Plan        string    `json:"plan"`
}

type contextKey string

const usernameKey contextKey = "preferred_username"

// authMiddleware проверяет JWT и кладет preferred_username в контекст
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

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "Invalid Authorization header", http.StatusUnauthorized)
			return
		}

		tokenStr := parts[1]

		// Парсим без проверки подписи
		token, _, err := jwt.NewParser().ParseUnverified(tokenStr, jwt.MapClaims{})
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			username, ok := claims["preferred_username"].(string)
			if !ok {
				http.Error(w, "preferred_username not found", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), usernameKey, username)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
	})
}

// helper для получения username из контекста
func getUsernameFromContext(ctx context.Context) string {
	if username, ok := ctx.Value(usernameKey).(string); ok {
		return username
	}
	return ""
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
	db, err := sqlx.Open("clickhouse", "clickhouse://clickhouse:9000")
	if err != nil {
		log.Fatal("Failed to connect to ClickHouse:", err)
	}

	reportsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username := getUsernameFromContext(r.Context()) // достаём username из контекста
		if username == "" {
			w.Header().Set("Content-Type", "application/json")
			return
		}

		query := db.Rebind(`
					SELECT
						user_id,
						sensor_value,
						ts,
						plan
					FROM user_reports
					WHERE user_id = ?
					ORDER BY ts DESC
					LIMIT 10
				`)
		rows, err := db.QueryContext(r.Context(), query, username)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var result []ReportRow
		for rows.Next() {
			var r ReportRow
			if err := rows.Scan(
				&r.UserID,
				&r.SensorValue,
				&r.Ts,
				&r.Plan,
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
