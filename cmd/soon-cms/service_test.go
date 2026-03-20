package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLatestMigrationVersion(t *testing.T) {
	dir := t.TempDir()

	files := []string{
		"0001_init_schema.up.sql",
		"0001_init_schema.down.sql",
		"0002_add_links.up.sql",
		"0002_add_links.down.sql",
		"0003_add_roadmap.up.sql",
		"0003_add_roadmap.down.sql",
	}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(dir, f), []byte("-- test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	version, err := latestMigrationVersion(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if version != 3 {
		t.Errorf("expected version 3, got %d", version)
	}
}

func TestLatestMigrationVersion_Empty(t *testing.T) {
	dir := t.TempDir()

	_, err := latestMigrationVersion(dir)
	if err == nil {
		t.Error("expected error for empty directory")
	}
}

func TestLatestMigrationVersion_InvalidDir(t *testing.T) {
	_, err := latestMigrationVersion("/nonexistent/path")
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}
