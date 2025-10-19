package postgres

import (
    "database/sql"
    "github.com/rodolfodpk/instagrano/internal/domain"
)

type UserRepository interface {
    Create(user *domain.User) error
    FindByEmail(email string) (*domain.User, error)
}

type postgresUserRepository struct {
    db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
    return &postgresUserRepository{db: db}
}

func (r *postgresUserRepository) Create(user *domain.User) error {
    query := `
        INSERT INTO users (username, email, password)
        VALUES ($1, $2, $3)
        RETURNING id, created_at, updated_at`
    return r.db.QueryRow(query, user.Username, user.Email, user.Password).
        Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *postgresUserRepository) FindByEmail(email string) (*domain.User, error) {
    user := &domain.User{}
    query := `SELECT id, username, email, password, created_at, updated_at FROM users WHERE email = $1`
    err := r.db.QueryRow(query, email).Scan(
        &user.ID, &user.Username, &user.Email, &user.Password,
        &user.CreatedAt, &user.UpdatedAt,
    )
    return user, err
}
