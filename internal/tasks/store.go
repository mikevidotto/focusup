package tasks

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func datadir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	appdir := filepath.Join(dir, "focusup")
	if err := os.MkdirAll(appdir, 0o755); err != nil {
		return "", err
	}

	return appdir, nil
}

func load[T any](path string) ([]T, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return []T{}, nil
	}
	if err != nil {
		return nil, err
	}

	var loaded []T
	if err := json.Unmarshal(data, &loaded); err != nil {
		return nil, err
	}
	if loaded == nil {
		loaded = []T{}
	}

	return loaded, nil
}

func save[T any](path string, items []T) error {
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}

	return os.Rename(tmp, path)
}
