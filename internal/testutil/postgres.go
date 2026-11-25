//go:build integration

// Package testutil gives integration tests a throwaway PostgreSQL database.
package testutil

import (
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/tanawat011/golang-boilerplate/internal/model"
)

// NewDatabase creates an empty, migrated database for one test and drops it
// afterwards. TEST_DATABASE_URL must point at a PostgreSQL server where the
// user may CREATE DATABASE, e.g.
//
//	postgres://postgres:postgres@127.0.0.1:15432/postgres?sslmode=disable
func NewDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	adminURL := os.Getenv("TEST_DATABASE_URL")
	if adminURL == "" {
		t.Fatal("TEST_DATABASE_URL is not set (see README, Testing)")
	}

	admin, err := gorm.Open(postgres.Open(adminURL), &gorm.Config{Logger: gormlogger.Discard})
	if err != nil {
		t.Fatalf("connect to %s: %v", redact(adminURL), err)
	}
	adminSQL, _ := admin.DB()
	t.Cleanup(func() { _ = adminSQL.Close() })

	name := fmt.Sprintf("gb_test_%d", time.Now().UnixNano())
	if err := admin.Exec("CREATE DATABASE " + name).Error; err != nil {
		t.Fatalf("create database: %v", err)
	}

	u, err := url.Parse(adminURL)
	if err != nil {
		t.Fatal(err)
	}
	u.Path = "/" + name
	db, err := gorm.Open(postgres.Open(u.String()), &gorm.Config{
		TranslateError: true,
		Logger:         gormlogger.Discard,
	})
	if err != nil {
		t.Fatalf("connect to test database: %v", err)
	}
	sqlDB, _ := db.DB()

	t.Cleanup(func() {
		_ = sqlDB.Close()
		if err := admin.Exec("DROP DATABASE IF EXISTS " + name + " WITH (FORCE)").Error; err != nil {
			t.Errorf("drop %s: %v", name, err)
		}
	})

	if err := db.AutoMigrate(&model.User{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func redact(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return "<invalid url>"
	}
	return u.Redacted()
}
