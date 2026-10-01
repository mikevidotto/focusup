package habits 

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

	return filepath.Join(appdir, "habits.json"), nil
}

func load(path string) ([]Habit, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return []Habit{}, nil
	}
	if err != nil {
		return nil, err
	}

	var loaded []Habit
	if err := json.Unmarshal(data, &loaded); err != nil {
		return nil, err
	}

	return loaded, nil
}

func save(path string, habits []Habit) error {
	data, err := json.MarshalIndent(habits, "", "  ")
	if err != nil {
		return err
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}

	return os.Rename(tmp, path)
}
