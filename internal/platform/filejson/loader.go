package filejson

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ValidateFunc validates a decoded manifest loaded from disk.
type ValidateFunc[T any] func(path string, value T) error

// RequirePath ensures the configured repository path is present.
func RequirePath(root string, subject string) error {
	if strings.TrimSpace(root) == "" {
		return fmt.Errorf("%s repository path is required", subject)
	}

	return nil
}

// LoadDirectory reads every JSON file in the directory and validates each decoded manifest.
func LoadDirectory[T any](root string, subject string, validate ValidateFunc[T]) ([]T, error) {
	if err := RequirePath(root, subject); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("read %s repository: %w", subject, err)
	}

	values := make([]T, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		value, err := readFile[T](filepath.Join(root, entry.Name()), validate)
		if err != nil {
			return nil, err
		}

		values = append(values, value)
	}

	return values, nil
}

func readFile[T any](path string, validate ValidateFunc[T]) (T, error) {
	file, err := os.Open(path)
	if err != nil {
		return zero[T](), fmt.Errorf("open manifest %q: %w", path, err)
	}
	defer file.Close()

	payload, err := io.ReadAll(file)
	if err != nil {
		return zero[T](), fmt.Errorf("read manifest %q: %w", path, err)
	}

	var value T
	if err := json.Unmarshal(payload, &value); err != nil {
		return zero[T](), fmt.Errorf("decode manifest %q: %w", path, err)
	}

	if validate == nil {
		return value, nil
	}

	if err := validate(path, value); err != nil {
		return zero[T](), err
	}

	return value, nil
}

func zero[T any]() T {
	var value T
	return value
}
