package workouts

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func datafilepath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	appdir := filepath.Join(dir, "focusup")
	if err := os.MkdirAll(appdir, 0o755); err != nil {
		return "", err
	}

	return filepath.Join(appdir, "workouts.json"), nil
}

func load(path string) ([]Cycle, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return []Cycle{}, nil
	}
	if err != nil {
		return nil, err
	}

	var loaded []Cycle
	if err := json.Unmarshal(data, &loaded); err != nil {
		return nil, err
	}

	return loaded, nil
}

func save(path string, cycles []Cycle) error {
	data, err := json.MarshalIndent(cycles, "", "  ")
	if err != nil {
		return err
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}

	return os.Rename(tmp, path)
}
