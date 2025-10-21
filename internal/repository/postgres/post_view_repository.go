package postgres

import (
	"database/sql"
	"time"

	"github.com/rodolfodpk/instagrano/internal/domain"
)

type PostViewRepository interface {
	StartView(view *domain.PostView) error
	EndView(userID, postID uint, startedAt, endedAt time.Time) error
	IncrementPostViewsCount(postID uint) error
}

type postgresPostViewRepository struct {
	db *sql.DB
}

func NewPostViewRepository(db *sql.DB) PostViewRepository {
	return &postgresPostViewRepository{db: db}
}

func (r *postgresPostViewRepository) StartView(view *domain.PostView) error {
	query := `INSERT INTO post_views (user_id, post_id, started_at) VALUES ($1, $2, $3) RETURNING id`
	return r.db.QueryRow(query, view.UserID, view.PostID, view.StartedAt).Scan(&view.ID)
}

func (r *postgresPostViewRepository) EndView(userID, postID uint, startedAt, endedAt time.Time) error {
	duration := int(endedAt.Sub(startedAt).Seconds())
	query := `
		UPDATE post_views 
		SET ended_at = $1, duration_seconds = $2 
		WHERE user_id = $3 AND post_id = $4 AND started_at = $5`
	_, err := r.db.Exec(query, endedAt, duration, userID, postID, startedAt)
	return err
}

func (r *postgresPostViewRepository) IncrementPostViewsCount(postID uint) error {
	query := `UPDATE posts SET views_count = views_count + 1 WHERE id = $1`
	_, err := r.db.Exec(query, postID)
	return err
}
