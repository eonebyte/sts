package auth

import (
	"context"
	"database/sql"
	"log"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	FindUser(ctx context.Context, username string) (*User, error)
	CreateUser(ctx context.Context, username, passwordHash string) error
}

type oraRepo struct {
	db *sqlx.DB
}

func NewOraRepository(db *sqlx.DB) Repository {
	return &oraRepo{db: db}
}

func (r *oraRepo) FindUser(ctx context.Context, username string) (*User, error) {
	var u User

	query := `
		SELECT AD_User_ID, Name, Password, Title
		FROM AD_User
		WHERE Name = :1
	`

	err := r.db.GetContext(ctx, &u, query, username)

	if err == sql.ErrNoRows {
		log.Printf("User not found: %s", username)
		return nil, nil
	}

	if err != nil {
		log.Printf("Error querying user %s: %v", username, err)
		return nil, err
	}

	return &u, nil
}

func (r *oraRepo) CreateUser(ctx context.Context, username, passwordHash string) error {
	query := `
		INSERT INTO users (username, hashed_password)
		VALUES ($1, $2)
	`

	_, err := r.db.ExecContext(ctx, query, username, passwordHash)
	if err != nil {
		log.Printf("Failed to create user '%s': %v", username, err)
		return err
	}

	log.Printf("New user created: %s", username)
	return nil
}
