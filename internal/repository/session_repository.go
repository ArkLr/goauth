package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Arklr/goauth/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionRepository struct {
	db *pgxpool.Pool
}

func NewSessionRepository(db *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, s *models.Session) error {
	const q = `
		INSERT INTO sessions (token, user_id, csrf_token, expires_at)
		VALUES ($1, $2, $3, $4)`

	_, err := r.db.Exec(ctx, q, s.Token, s.UserId, s.CSRFToken, s.ExpiresAt)
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}

func (r *SessionRepository) GetByToken(ctx context.Context, token string) (*models.Session, error) {
	const q = `
		SELECT token, user_id, csrf_token, expires_at
		FROM sessions
		WHERE token = $1 AND expires_at > now()`

	var s models.Session
	err := r.db.QueryRow(ctx, q, token).
		Scan(&s.Token, &s.UserId, &s.CSRFToken, &s.ExpiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get session: %w", err)
	}
	return &s, nil
}

func (r *SessionRepository) Delete(ctx context.Context, token string) error {
	const q = `DELETE FROM sessions WHERE token = $1`

	if _, err := r.db.Exec(ctx, q, token); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}
