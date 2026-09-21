package posts

import (
	"context"
	"errors"
	"strings"
	"time"
)

const (
	MaxTitleLength   = 200
	MaxContentLength = 10_000
)

var ErrNotFound = errors.New("post not found")

type Post struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Input struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (input *Input) Validate() string {
	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" {
		return "title is required"
	}
	if len([]rune(input.Title)) > MaxTitleLength {
		return "title must be 200 characters or fewer"
	}
	if len([]rune(input.Content)) > MaxContentLength {
		return "content must be 10000 characters or fewer"
	}
	return ""
}

type Store interface {
	List(context.Context) ([]Post, error)
	Get(context.Context, int64) (Post, error)
	Create(context.Context, Input) (Post, error)
	Update(context.Context, int64, Input) (Post, error)
	Delete(context.Context, int64) error
}
