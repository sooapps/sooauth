package store

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Audit struct {
	db *pgxpool.Pool
}

func NewAudit(db *pgxpool.Pool) *Audit {
	return &Audit{db: db}
}

func (a *Audit) Log(ctx context.Context, userID *uuid.UUID, action string, meta map[string]any, ip string) {
	payload, _ := json.Marshal(meta)
	var uid any
	if userID != nil {
		uid = *userID
	}
	_, _ = a.db.Exec(ctx, `
		INSERT INTO audit_log (user_id, action, meta, ip, created_at)
		VALUES ($1, $2, $3::jsonb, NULLIF($4, '')::inet, now())
	`, uid, action, string(payload), ip)
}
