package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/KouroshEsmaeili/go-rest-api-template/internal/posts"
)

type memoryStore struct {
	mu     sync.Mutex
	nextID int64
	items  map[int64]posts.Post
	err    error
}

func newMemoryStore() *memoryStore {
	return &memoryStore{nextID: 1, items: make(map[int64]posts.Post)}
}

func (store *memoryStore) List(context.Context) ([]posts.Post, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	if store.err != nil {
		return nil, store.err
	}
	result := make([]posts.Post, 0, len(store.items))
	for _, post := range store.items {
		result = append(result, post)
	}
	return result, nil
}

func (store *memoryStore) Get(_ context.Context, id int64) (posts.Post, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	if store.err != nil {
		return posts.Post{}, store.err
	}
	post, ok := store.items[id]
	if !ok {
		return posts.Post{}, posts.ErrNotFound
	}
	return post, nil
}

func (store *memoryStore) Create(_ context.Context, input posts.Input) (posts.Post, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	if store.err != nil {
		return posts.Post{}, store.err
	}
	now := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	post := posts.Post{ID: store.nextID, Title: input.Title, Content: input.Content, CreatedAt: now, UpdatedAt: now}
	store.items[post.ID] = post
	store.nextID++
	return post, nil
}

func (store *memoryStore) Update(_ context.Context, id int64, input posts.Input) (posts.Post, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	if store.err != nil {
		return posts.Post{}, store.err
	}
	post, ok := store.items[id]
	if !ok {
		return posts.Post{}, posts.ErrNotFound
	}
	post.Title = input.Title
	post.Content = input.Content
	post.UpdatedAt = post.UpdatedAt.Add(time.Second)
	store.items[id] = post
	return post, nil
}

func (store *memoryStore) Delete(_ context.Context, id int64) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	if store.err != nil {
		return store.err
	}
	if _, ok := store.items[id]; !ok {
		return posts.ErrNotFound
	}
	delete(store.items, id)
	return nil
}

type stubPinger struct{ err error }

func (pinger stubPinger) PingContext(context.Context) error { return pinger.err }

func testRouter(store posts.Store, pinger stubPinger) http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewRouter("test", store, pinger, logger)
}

func performRequest(handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func TestHealthEndpoints(t *testing.T) {
	store := newMemoryStore()

	t.Run("live", func(t *testing.T) {
		response := performRequest(testRouter(store, stubPinger{}), http.MethodGet, "/health", "")
		if response.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
		}
	})

	t.Run("ready when database is unavailable", func(t *testing.T) {
		response := performRequest(testRouter(store, stubPinger{err: errors.New("unavailable")}), http.MethodGet, "/ready", "")
		if response.Code != http.StatusServiceUnavailable {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
		}
	})
}

func TestCreateRejectsMalformedJSON(t *testing.T) {
	response := performRequest(testRouter(newMemoryStore(), stubPinger{}), http.MethodPost, "/api/v1/posts", `{"title":`)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	assertErrorCode(t, response, "invalid_json")
}

func TestCreateRejectsInvalidInput(t *testing.T) {
	response := performRequest(testRouter(newMemoryStore(), stubPinger{}), http.MethodPost, "/api/v1/posts", `{"title":"  "}`)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	assertErrorCode(t, response, "validation_error")
}

func TestPostLifecycle(t *testing.T) {
	router := testRouter(newMemoryStore(), stubPinger{})

	created := performRequest(router, http.MethodPost, "/api/v1/posts", `{"title":"First post","content":"Hello"}`)
	if created.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d; body = %s", created.Code, http.StatusCreated, created.Body.String())
	}
	var createBody struct {
		Data posts.Post `json:"data"`
	}
	decodeResponse(t, created, &createBody)
	if createBody.Data.ID != 1 || createBody.Data.Title != "First post" {
		t.Fatalf("created post = %+v", createBody.Data)
	}

	fetched := performRequest(router, http.MethodGet, "/api/v1/posts/1", "")
	if fetched.Code != http.StatusOK {
		t.Fatalf("get status = %d, want %d", fetched.Code, http.StatusOK)
	}
	var getBody struct {
		Data posts.Post `json:"data"`
	}
	decodeResponse(t, fetched, &getBody)
	if getBody.Data.ID != 1 || getBody.Data.Title != "First post" {
		t.Fatalf("retrieved post = %+v", getBody.Data)
	}

	updated := performRequest(router, http.MethodPut, "/api/v1/posts/1", `{"title":"Updated","content":"New content"}`)
	if updated.Code != http.StatusOK {
		t.Fatalf("update status = %d, want %d; body = %s", updated.Code, http.StatusOK, updated.Body.String())
	}
	var updateBody struct {
		Data posts.Post `json:"data"`
	}
	decodeResponse(t, updated, &updateBody)
	if updateBody.Data.Title != "Updated" || updateBody.Data.Content != "New content" {
		t.Fatalf("updated post = %+v", updateBody.Data)
	}

	deleted := performRequest(router, http.MethodDelete, "/api/v1/posts/1", "")
	if deleted.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d, want %d", deleted.Code, http.StatusNoContent)
	}

	missing := performRequest(router, http.MethodGet, "/api/v1/posts/1", "")
	if missing.Code != http.StatusNotFound {
		t.Fatalf("missing status = %d, want %d", missing.Code, http.StatusNotFound)
	}
	assertErrorCode(t, missing, "not_found")
}

func TestInvalidID(t *testing.T) {
	response := performRequest(testRouter(newMemoryStore(), stubPinger{}), http.MethodGet, "/api/v1/posts/not-a-number", "")
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	assertErrorCode(t, response, "invalid_id")
}

func TestDatabaseErrorDoesNotLeak(t *testing.T) {
	store := newMemoryStore()
	store.err = errors.New("password=secret database exploded")
	response := performRequest(testRouter(store, stubPinger{}), http.MethodGet, "/api/v1/posts", "")
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusInternalServerError)
	}
	if bytes.Contains(response.Body.Bytes(), []byte("secret")) {
		t.Fatalf("response leaked internal error: %s", response.Body.String())
	}
	assertErrorCode(t, response, "internal_error")
}

func assertErrorCode(t *testing.T, response *httptest.ResponseRecorder, want string) {
	t.Helper()
	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	decodeResponse(t, response, &body)
	if body.Error.Code != want {
		t.Fatalf("error code = %q, want %q", body.Error.Code, want)
	}
}

func decodeResponse(t *testing.T, response *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.Unmarshal(response.Body.Bytes(), target); err != nil {
		t.Fatalf("decode response: %v; body = %s", err, response.Body.String())
	}
}
