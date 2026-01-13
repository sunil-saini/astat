package cache

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

func Dir() string {
	if d := viper.GetString("cache_dir"); d != "" {
		return d
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(
		home,
		".cache",
		"astat",
	)
}
