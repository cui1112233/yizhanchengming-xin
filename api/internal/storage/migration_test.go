package storage

import (
	"os"
	"strings"
	"testing"
)

func TestInitialMigrationContainsUnifiedBatchPipelineTables(t *testing.T) {
	body, err := os.ReadFile("../../migrations/00001_init.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := string(body)
	for _, table := range []string{
		"intakes",
		"intake_books",
		"batches",
		"batch_books",
		"pipeline_jobs",
		"pipeline_stages",
	} {
		if !strings.Contains(sql, "CREATE TABLE "+table) && !strings.Contains(sql, "CREATE TABLE IF NOT EXISTS "+table) {
			t.Fatalf("migration missing table %s", table)
		}
	}
	for _, required := range []string{"run_at", "gender_source", "category", "genre", "retry_count", "error_message"} {
		if !strings.Contains(sql, required) {
			t.Fatalf("migration missing required field %s", required)
		}
	}
}

func TestInitialMigrationHasGooseUpAndDownSections(t *testing.T) {
	body, err := os.ReadFile("../../migrations/00001_init.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := string(body)
	if !strings.Contains(sql, "-- +goose Up") || !strings.Contains(sql, "-- +goose Down") {
		t.Fatal("migration must be Goose compatible")
	}
}
