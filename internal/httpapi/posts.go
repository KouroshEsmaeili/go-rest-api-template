package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/KouroshEsmaeili/go-rest-api-template/internal/posts"
	"github.com/gin-gonic/gin"
)

const maxRequestBodySize = 1 << 20

type PostHandler struct {
	store  posts.Store
	logger *slog.Logger
}

func NewPostHandler(store posts.Store, logger *slog.Logger) *PostHandler {
	return &PostHandler{store: store, logger: logger}
}

func (handler *PostHandler) List(c *gin.Context) {
	result, err := handler.store.List(c.Request.Context())
	if err != nil {
		handler.internalError(c, "list posts", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (handler *PostHandler) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	post, err := handler.store.Get(c.Request.Context(), id)
	if err != nil {
		handler.storeError(c, "get post", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": post})
}

func (handler *PostHandler) Create(c *gin.Context) {
	input, ok := decodeInput(c)
	if !ok {
		return
	}
	post, err := handler.store.Create(c.Request.Context(), input)
	if err != nil {
		handler.internalError(c, "create post", err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": post})
}

func (handler *PostHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	input, ok := decodeInput(c)
	if !ok {
		return
	}
	post, err := handler.store.Update(c.Request.Context(), id, input)
	if err != nil {
		handler.storeError(c, "update post", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": post})
}

func (handler *PostHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := handler.store.Delete(c.Request.Context(), id); err != nil {
		handler.storeError(c, "delete post", err)
		return
	}
	c.Status(http.StatusNoContent)
}

func decodeInput(c *gin.Context) (posts.Input, bool) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestBodySize)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()

	var input posts.Input
	if err := decoder.Decode(&input); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_json", "request body must contain one valid JSON object")
		return posts.Input{}, false
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeError(c, http.StatusBadRequest, "invalid_json", "request body must contain one valid JSON object")
		return posts.Input{}, false
	}
	if message := input.Validate(); message != "" {
		writeError(c, http.StatusBadRequest, "validation_error", message)
		return posts.Input{}, false
	}
	return input, true
}

func parseID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		writeError(c, http.StatusBadRequest, "invalid_id", "id must be a positive integer")
		return 0, false
	}
	return id, true
}

func (handler *PostHandler) storeError(c *gin.Context, operation string, err error) {
	if errors.Is(err, posts.ErrNotFound) {
		writeError(c, http.StatusNotFound, "not_found", "post not found")
		return
	}
	handler.internalError(c, operation, err)
}

func (handler *PostHandler) internalError(c *gin.Context, operation string, err error) {
	handler.logger.Error("request failed", "operation", operation, "error", err)
	writeError(c, http.StatusInternalServerError, "internal_error", "an unexpected error occurred")
}

func writeError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	})
}
