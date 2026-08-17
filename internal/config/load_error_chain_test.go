package config

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// TestLoadMissingFileIsIdentifiableAsNotExist asserts that a missing store
// configuration file is reported as a "file does not exist" failure that a
// caller can classify with errors.Is / errors.As, without comparing message
// text.
func TestLoadMissingFileIsIdentifiableAsNotExist(t *testing.T) {
	path := filepath.Join(t.TempDir(), "no-such-config.json")

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected an error for a missing config file, got nil")
	}
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("a missing config file must be identifiable as fs.ErrNotExist, got %v", err)
	}
	var pathErr *fs.PathError
	if !errors.As(err, &pathErr) {
		t.Errorf("a missing config file error must unwrap to *fs.PathError, got %v", err)
	} else if pathErr.Path != path {
		t.Errorf("path error path = %q, want %q", pathErr.Path, path)
	}
}

// TestLoadDirectoryIsIdentifiableAsPathError asserts the same classification
// works for the other unreadable-path case.
func TestLoadDirectoryIsIdentifiableAsPathError(t *testing.T) {
	dir := t.TempDir()

	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected an error when the config path is a directory, got nil")
	}
	var pathErr *fs.PathError
	if !errors.As(err, &pathErr) {
		t.Errorf("an unreadable config path error must unwrap to *fs.PathError, got %v", err)
	}
	if errors.Is(err, fs.ErrNotExist) {
		t.Errorf("an existing directory must not be reported as fs.ErrNotExist, got %v", err)
	}
}

// TestLoadInvalidContentIsNotNotExist is the contrast case: a readable file
// with invalid content must never look like a missing file, and its JSON syntax
// failure must stay classifiable.
func TestLoadInvalidContentIsNotNotExist(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{not json`), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected a parse error, got nil")
	}
	if errors.Is(err, fs.ErrNotExist) {
		t.Errorf("invalid config content must not be reported as fs.ErrNotExist, got %v", err)
	}
	var syntaxErr *json.SyntaxError
	if !errors.As(err, &syntaxErr) {
		t.Errorf("invalid config JSON must unwrap to *json.SyntaxError, got %v", err)
	}
}
