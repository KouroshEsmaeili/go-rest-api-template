package posts

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (store *PostgresStore) List(ctx context.Context) ([]Post, error) {
	rows, err := store.db.QueryContext(ctx, `
		SELECT id, title, content, created_at, updated_at
		FROM posts
		ORDER BY id DESC`)
	if err != nil {
		return nil, fmt.Errorf("list posts: %w", err)
	}
	defer rows.Close()

	result := make([]Post, 0)
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.CreatedAt, &post.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan post: %w", err)
		}
		result = append(result, post)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate posts: %w", err)
	}
	return result, nil
}

func (store *PostgresStore) Get(ctx context.Context, id int64) (Post, error) {
	var post Post
	err := store.db.QueryRowContext(ctx, `
		SELECT id, title, content, created_at, updated_at
		FROM posts
		WHERE id = $1`, id,
	).Scan(&post.ID, &post.Title, &post.Content, &post.CreatedAt, &post.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Post{}, ErrNotFound
	}
	if err != nil {
		return Post{}, fmt.Errorf("get post: %w", err)
	}
	return post, nil
}

func (store *PostgresStore) Create(ctx context.Context, input Input) (Post, error) {
	var post Post
	err := store.db.QueryRowContext(ctx, `
		INSERT INTO posts (title, content)
		VALUES ($1, $2)
		RETURNING id, title, content, created_at, updated_at`,
		input.Title, input.Content,
	).Scan(&post.ID, &post.Title, &post.Content, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		return Post{}, fmt.Errorf("create post: %w", err)
	}
	return post, nil
}

func (store *PostgresStore) Update(ctx context.Context, id int64, input Input) (Post, error) {
	var post Post
	err := store.db.QueryRowContext(ctx, `
		UPDATE posts
		SET title = $2, content = $3, updated_at = NOW()
		WHERE id = $1
		RETURNING id, title, content, created_at, updated_at`,
		id, input.Title, input.Content,
	).Scan(&post.ID, &post.Title, &post.Content, &post.CreatedAt, &post.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Post{}, ErrNotFound
	}
	if err != nil {
		return Post{}, fmt.Errorf("update post: %w", err)
	}
	return post, nil
}

func (store *PostgresStore) Delete(ctx context.Context, id int64) error {
	result, err := store.db.ExecContext(ctx, `DELETE FROM posts WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete post: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read deleted row count: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}
