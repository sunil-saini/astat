package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func Write(path string, v any) error {
	tmp := path + ".tmp"

	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}

	return os.Rename(tmp, path)
}

func Read(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func EnsureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}

func Path(cacheDir, name string) string {
	return filepath.Join(cacheDir, name+".json")
}
