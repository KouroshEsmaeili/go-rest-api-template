package config

import (
	"os"
	"testing"
)

func TestLoadUsesEnvironment(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("APP_ENV", "test")
	t.Setenv("DATABASE_URL", "postgres://example")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Port != "9090" || cfg.Environment != "test" || cfg.DatabaseURL != "postgres://example" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func TestLoadRejectsInvalidPort(t *testing.T) {
	t.Setenv("PORT", "70000")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want invalid port error")
	}
}

func TestLoadRejectsEmptyDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want empty database URL error")
	}
}

func TestLoadRequiresDatabaseURLInProductionWhenUnset(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("PORT", "8080")
	unsetEnv(t, "DATABASE_URL")

	if _, err := Load(); err == nil || err.Error() != "DATABASE_URL must be set in production" {
		t.Fatalf("Load() error = %v, want DATABASE_URL production requirement error", err)
	}
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()

	value, wasSet := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("unset %s: %v", key, err)
	}

	t.Cleanup(func() {
		var err error
		if wasSet {
			err = os.Setenv(key, value)
		} else {
			err = os.Unsetenv(key)
		}
		if err != nil {
			t.Errorf("restore %s: %v", key, err)
		}
	})
}
