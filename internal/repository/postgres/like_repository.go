package postgres

import (
	"database/sql"
	"fmt"

	"github.com/rodolfodpk/instagrano/internal/domain"
)

type LikeRepository interface {
	Create(like *domain.Like) error
	IncrementPostLikeCount(postID uint) error
	DecrementPostLikeCount(postID uint) error
	FindByPostID(postID uint) ([]*domain.Like, error)
	FindByUserAndPost(userID, postID uint) (*domain.Like, error)
	Delete(userID, postID uint) error
}

type CommentRepository interface {
	Create(comment *domain.Comment) error
	IncrementPostCommentCount(postID uint) error
	FindByPostID(postID uint) ([]*domain.Comment, error)
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

func (r *postgresLikeRepository) DecrementPostLikeCount(postID uint) error {
	query := `UPDATE posts SET likes_count = likes_count - 1 WHERE id = $1`
	_, err := r.db.Exec(query, postID)
	return err
}

func (r *postgresLikeRepository) FindByUserAndPost(userID, postID uint) (*domain.Like, error) {
	query := `SELECT id, user_id, post_id, created_at FROM likes WHERE user_id = $1 AND post_id = $2`
	like := &domain.Like{}
	err := r.db.QueryRow(query, userID, postID).Scan(&like.ID, &like.UserID, &like.PostID, &like.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found, but not an error
		}
		return nil, err
	}
	return like, nil
}

func (r *postgresLikeRepository) Delete(userID, postID uint) error {
	query := `DELETE FROM likes WHERE user_id = $1 AND post_id = $2`
	result, err := r.db.Exec(query, userID, postID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("like not found")
	}
	return nil
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

func (r *postgresLikeRepository) FindByPostID(postID uint) ([]*domain.Like, error) {
	query := `SELECT id, user_id, post_id, created_at FROM likes WHERE post_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var likes []*domain.Like
	for rows.Next() {
		like := &domain.Like{}
		err := rows.Scan(&like.ID, &like.UserID, &like.PostID, &like.CreatedAt)
		if err != nil {
			return nil, err
		}
		likes = append(likes, like)
	}
	return likes, nil
}

func (r *postgresCommentRepository) FindByPostID(postID uint) ([]*domain.Comment, error) {
	query := `
		SELECT c.id, c.user_id, c.post_id, c.text, c.created_at, u.username
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.post_id = $1 
		ORDER BY c.created_at ASC
	`
	rows, err := r.db.Query(query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*domain.Comment
	for rows.Next() {
		comment := &domain.Comment{}
		err := rows.Scan(&comment.ID, &comment.UserID, &comment.PostID, &comment.Text, &comment.CreatedAt, &comment.Username)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	return comments, nil
}
