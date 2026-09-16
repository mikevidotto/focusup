package tasks

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func dataFilePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	appDir := filepath.Join(dir, "focusup")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		return "", err
	}

	return filepath.Join(appDir, "tasks.json"), nil
}

func load(path string) ([]Task, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return []Task{}, nil
	}
	if err != nil {
		return nil, err
	}

	var loaded []Task
	if err := json.Unmarshal(data, &loaded); err != nil {
		return nil, err
	}

	return loaded, nil
}

func save(path string, tasks []Task) error {
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return err
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}

	return os.Rename(tmp, path)
}
