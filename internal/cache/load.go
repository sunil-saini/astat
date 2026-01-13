package cache

import (
	"os"
)

func Load(path string, v any) (bool, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	if err := Read(path, v); err != nil {
		return false, err
	}

	return true, nil
}
