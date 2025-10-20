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
	GetFeed(limit, offset int) ([]*domain.Post, error)
	GetFeedWithCursor(limit int, cursor *pagination.Cursor) ([]*domain.Post, error)
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
		SELECT id, user_id, title, caption, media_type, media_url,
			   likes_count, comments_count, views_count, created_at, updated_at
		FROM posts WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&post.ID, &post.UserID, &post.Title, &post.Caption, &post.MediaType,
		&post.MediaURL, &post.LikesCount, &post.CommentsCount, &post.ViewsCount,
		&post.CreatedAt, &post.UpdatedAt,
	)
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
			&post.ID, &post.UserID, &post.Title, &post.Caption, &post.MediaType,
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
			SELECT id, user_id, title, caption, media_type, media_url,
				   likes_count, comments_count, views_count, created_at, updated_at
			FROM posts 
			ORDER BY created_at DESC, id DESC 
			LIMIT $1`
		args = []interface{}{limit}
	} else {
		// Subsequent pages - use cursor
		query = `
			SELECT id, user_id, title, caption, media_type, media_url,
				   likes_count, comments_count, views_count, created_at, updated_at
			FROM posts 
			WHERE (created_at < $2) OR (created_at = $2 AND id < $3)
			ORDER BY created_at DESC, id DESC 
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
			&post.ID, &post.UserID, &post.Title, &post.Caption, &post.MediaType,
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
