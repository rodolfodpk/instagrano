package postgres

import (
	"database/sql"
	"fmt"

	"github.com/rodolfodpk/instagrano/internal/domain"
	"github.com/rodolfodpk/instagrano/internal/pagination"
)

type PostRepository interface {
	Create(post *domain.Post) error
	FindByID(id uint) (*domain.Post, error)
	GetByID(id uint) (*domain.Post, error)
	GetFeed(limit, offset int) ([]*domain.Post, error)
	GetFeedWithCursor(limit int, cursor *pagination.Cursor) ([]*domain.Post, error)
	Delete(id uint) error
}

type postgresPostRepository struct {
	db *sql.DB
}

func NewPostRepository(db *sql.DB) PostRepository {
	return &postgresPostRepository{db: db}
}

func (r *postgresPostRepository) Create(post *domain.Post) error {
	query := `
		INSERT INTO posts (user_id, title, caption, media_type, media_url)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`
	return r.db.QueryRow(query, post.UserID, post.Title, post.Caption,
		post.MediaType, post.MediaURL).
		Scan(&post.ID, &post.CreatedAt, &post.UpdatedAt)
}

func (r *postgresPostRepository) FindByID(id uint) (*domain.Post, error) {
	post := &domain.Post{}
	query := `
		SELECT p.id, p.user_id, u.username, p.title, p.caption, p.media_type, p.media_url,
			   p.likes_count, p.comments_count, p.views_count, p.created_at, p.updated_at
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&post.ID, &post.UserID, &post.Username, &post.Title, &post.Caption, &post.MediaType,
		&post.MediaURL, &post.LikesCount, &post.CommentsCount, &post.ViewsCount,
		&post.CreatedAt, &post.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("post with id %d not found", id)
	}
	return post, err
}

func (r *postgresPostRepository) GetFeed(limit, offset int) ([]*domain.Post, error) {
	query := `
		SELECT id, user_id, title, caption, media_type, media_url,
			   likes_count, comments_count, views_count, created_at, updated_at
		FROM posts ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*domain.Post
	for rows.Next() {
		post := &domain.Post{}
		err := rows.Scan(
			&post.ID, &post.UserID, &post.Username, &post.Title, &post.Caption, &post.MediaType,
			&post.MediaURL, &post.LikesCount, &post.CommentsCount, &post.ViewsCount,
			&post.CreatedAt, &post.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, nil
}

func (r *postgresPostRepository) GetFeedWithCursor(limit int, cursor *pagination.Cursor) ([]*domain.Post, error) {
	var query string
	var args []interface{}

	if cursor == nil {
		// First page - no cursor
		query = `
			SELECT p.id, p.user_id, u.username, p.title, p.caption, p.media_type, p.media_url,
				   p.likes_count, p.comments_count, p.views_count, p.created_at, p.updated_at
			FROM posts p
			JOIN users u ON p.user_id = u.id
			ORDER BY p.created_at DESC, p.id DESC 
			LIMIT $1`
		args = []interface{}{limit}
	} else {
		// Subsequent pages - use cursor
		query = `
			SELECT p.id, p.user_id, u.username, p.title, p.caption, p.media_type, p.media_url,
				   p.likes_count, p.comments_count, p.views_count, p.created_at, p.updated_at
			FROM posts p
			JOIN users u ON p.user_id = u.id
			WHERE (p.created_at < $2) OR (p.created_at = $2 AND p.id < $3)
			ORDER BY p.created_at DESC, p.id DESC 
			LIMIT $1`
		args = []interface{}{limit, cursor.Timestamp, cursor.ID}
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query feed with cursor: %w", err)
	}
	defer rows.Close()

	var posts []*domain.Post
	for rows.Next() {
		post := &domain.Post{}
		err := rows.Scan(
			&post.ID, &post.UserID, &post.Username, &post.Title, &post.Caption, &post.MediaType,
			&post.MediaURL, &post.LikesCount, &post.CommentsCount, &post.ViewsCount,
			&post.CreatedAt, &post.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}
		posts = append(posts, post)
	}
	return posts, nil
}

// GetByID gets a post by ID (alias for FindByID for consistency)
func (r *postgresPostRepository) GetByID(id uint) (*domain.Post, error) {
	return r.FindByID(id)
}

// Delete deletes a post by ID
func (r *postgresPostRepository) Delete(id uint) error {
	query := `DELETE FROM posts WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("post not found")
	}

	return nil
}
