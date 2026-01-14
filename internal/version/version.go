package version

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/sunil-saini/astat/internal/logger"
)

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

func GetLatestVersion() (string, string, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/sunil-saini/astat/releases/latest")
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch latest version: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", "", fmt.Errorf("failed to decode response: %w", err)
	}

	return release.TagName, release.HTMLURL, nil
}

func IsUpgradeAvailable() (bool, string, string, error) {
	if Version == "dev" {
		logger.Warn("dev version, skipping upgrade check")
		return false, "", "", nil
	}

	latestTag, url, err := GetLatestVersion()
	if err != nil {
		return false, "", "", err
	}

	current, err := version.NewVersion(Version)
	if err != nil {
		return false, "", "", fmt.Errorf("invalid current version: %w", err)
	}

	latest, err := version.NewVersion(latestTag)
	if err != nil {
		return false, "", "", fmt.Errorf("invalid latest version: %w", err)
	}

	return latest.GreaterThan(current), latestTag, url, nil
}
