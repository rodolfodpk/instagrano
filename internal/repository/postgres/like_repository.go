package postgres

import (
    "database/sql"
    "github.com/rodolfodpk/instagrano/internal/domain"
)

type LikeRepository interface {
    Create(like *domain.Like) error
    IncrementPostLikeCount(postID uint) error
}

type CommentRepository interface {
    Create(comment *domain.Comment) error
    IncrementPostCommentCount(postID uint) error
}

type postgresLikeRepository struct {
    db *sql.DB
}

type postgresCommentRepository struct {
    db *sql.DB
}

func NewLikeRepository(db *sql.DB) LikeRepository {
    return &postgresLikeRepository{db: db}
}

func NewCommentRepository(db *sql.DB) CommentRepository {
    return &postgresCommentRepository{db: db}
}

func (r *postgresLikeRepository) Create(like *domain.Like) error {
    query := `INSERT INTO likes (user_id, post_id) VALUES ($1, $2) RETURNING id, created_at`
    return r.db.QueryRow(query, like.UserID, like.PostID).Scan(&like.ID, &like.CreatedAt)
}

func (r *postgresLikeRepository) IncrementPostLikeCount(postID uint) error {
    query := `UPDATE posts SET likes_count = likes_count + 1 WHERE id = $1`
    _, err := r.db.Exec(query, postID)
    return err
}

func (r *postgresCommentRepository) Create(comment *domain.Comment) error {
    query := `INSERT INTO comments (user_id, post_id, text) VALUES ($1, $2, $3) RETURNING id, created_at`
    return r.db.QueryRow(query, comment.UserID, comment.PostID, comment.Text).Scan(&comment.ID, &comment.CreatedAt)
}

func (r *postgresCommentRepository) IncrementPostCommentCount(postID uint) error {
    query := `UPDATE posts SET comments_count = comments_count + 1 WHERE id = $1`
    _, err := r.db.Exec(query, postID)
    return err
}
