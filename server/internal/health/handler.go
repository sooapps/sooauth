package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Status struct {
	OK       bool              `json:"ok"`
	Checks   map[string]string `json:"checks,omitempty"`
	Database string            `json:"database,omitempty"`
	Redis    string            `json:"redis,omitempty"`
}

func Handler(db *pgxpool.Pool, rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		status := Status{OK: true, Checks: map[string]string{}}

		if db != nil {
			if err := db.Ping(ctx); err != nil {
				status.OK = false
				status.Checks["postgres"] = "down"
			} else {
				status.Checks["postgres"] = "up"
			}
		} else {
			status.Database = "not_configured"
		}

		if rdb != nil {
			if err := rdb.Ping(ctx).Err(); err != nil {
				status.OK = false
				status.Checks["redis"] = "down"
			} else {
				status.Checks["redis"] = "up"
			}
		} else {
			status.Redis = "not_configured"
		}

		code := http.StatusOK
		if !status.OK {
			code = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(status)
	}
}
