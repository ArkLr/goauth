package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Arklr/goauth/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostRepository struct {
	db *pgxpool.Pool
}

func NewPostRepository(db *pgxpool.Pool) *PostRepository {
	return &PostRepository{db: db}
}

const postCols = `id, user_id, headline, body, created_at, updated_at`

func scanPost(row pgx.Row) (*models.Post, error) {
	var p models.Post
	err := row.Scan(&p.Id, &p.UserId, &p.Headline, &p.Body, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *PostRepository) Create(ctx context.Context, userId int64, headline, body string) (*models.Post, error) {
	const q = `
		INSERT INTO posts (user_id, headline, body)
		VALUES ($1, $2, $3)
		RETURNING ` + postCols
	p, err := scanPost(r.db.QueryRow(ctx, q, userId, headline, body))
	if err != nil {
		return nil, fmt.Errorf("insert post: %w", err)

	}
	return p, nil
}

func (r *PostRepository) GetByID(ctx context.Context, id int64) (*models.Post, error) {
	const q = `SELECT ` + postCols + ` FROM posts WHERE id = $1`

	p, err := scanPost(r.db.QueryRow(ctx, q, id))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("get post: %w", err)
	}
	return p, nil
}

// List returns posts newest first along with their author's email.
// If userID is non-zero, only that user's posts are returned.
func (r *PostRepository) List(ctx context.Context, userID int64, limit, offset int) ([]models.Post, error) {
	const q = `
		SELECT p.id, p.user_id, p.headline, p.body, p.created_at, p.updated_at, u.email
		FROM posts p
		JOIN users u ON u.id = p.user_id
		WHERE $1::bigint = 0 OR p.user_id = $1::bigint
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, q, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list posts: %w", err)
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var p models.Post
		if err := rows.Scan(&p.Id, &p.UserId, &p.Headline, &p.Body, &p.CreatedAt, &p.UpdatedAt, &p.AuthorEmail); err != nil {
			return nil, fmt.Errorf("scan post: %w", err)
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func (r *PostRepository) CountByUser(ctx context.Context, userID int64) (int64, error) {
	const q = `SELECT count(*) FROM posts WHERE user_id = $1`

	var n int64
	if err := r.db.QueryRow(ctx, q, userID).Scan(&n); err != nil {
		return 0, fmt.Errorf("count posts: %w", err)
	}
	return n, nil
}

// Update changes a post only if it belongs to userID.
// If the post doesn't exist or isn't theirs, it returns ErrNotFound.
func (r *PostRepository) Update(ctx context.Context, id, userID int64, headline, body string) (*models.Post, error) {
	const q = `
		UPDATE posts
		SET headline = $3, body = $4, updated_at = now()
		WHERE id = $1 AND user_id = $2
		RETURNING ` + postCols

	p, err := scanPost(r.db.QueryRow(ctx, q, id, userID, headline, body))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("update post: %w", err)
	}
	return p, nil
}

func (r *PostRepository) Delete(ctx context.Context, id, userID int64) error {
	const q = `DELETE FROM posts WHERE id = $1 AND user_id = $2`

	tag, err := r.db.Exec(ctx, q, id, userID)
	if err != nil {
		return fmt.Errorf("delete post: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
