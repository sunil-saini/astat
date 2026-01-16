package cmd

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/sunil-saini/astat/internal/logger"
	"github.com/sunil-saini/astat/internal/version"
)

var upgradeCmd = &cobra.Command{
	Use:     "upgrade",
	Short:   "Upgrade astat to the latest version",
	GroupID: "project",
	Run: func(cmd *cobra.Command, args []string) {
		pterm.DefaultHeader.Println("Checking for updates...")
		fmt.Println()

		available, latestVersion, url, err := version.IsUpgradeAvailable()
		if err != nil {
			logger.Error("Failed to check for updates: %v", err)
			return
		}

		if !available {
			logger.Success("You are already on the latest version (%s)", version.Version)
			return
		}

		logger.Info("New version available: %s (current: %s)", latestVersion, version.Version)
		logger.Info("Download URL: %s", url)
		fmt.Println()

		downloadURL := getDownloadURL(latestVersion)
		if downloadURL == "" {
			logger.Error("Unsupported platform: %s/%s", runtime.GOOS, runtime.GOARCH)
			logger.Info("Please download manually from: %s", url)
			return
		}

		logger.Info("Downloading %s...", downloadURL)

		if err := downloadAndReplace(downloadURL); err != nil {
			logger.Error("Upgrade failed: %v", err)
			logger.Info("Please upgrade manually from: %s", url)
			return
		}

		logger.Success("Successfully upgraded to version %s!", latestVersion)
		logger.Info("Run 'astat version' to verify the upgrade")
	},
}

func getDownloadURL(version string) string {
	baseURL := fmt.Sprintf("https://github.com/sunil-saini/astat/releases/download/%s/astat_%s", version, version[1:])

	// Use lowercase OS names to match GoReleaser output
	platform := runtime.GOOS // darwin or linux

	var arch string
	switch runtime.GOARCH {
	case "amd64":
		arch = "x86_64"
	case "arm64":
		arch = "arm64"
	default:
		return ""
	}

	return fmt.Sprintf("%s_%s_%s.tar.gz", baseURL, platform, arch)
}

func downloadAndReplace(url string) error {
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	tmpFile, err := os.CreateTemp("", "astat-upgrade-*.tar.gz")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return fmt.Errorf("failed to save download: %w", err)
	}
	tmpFile.Close()

	binaryPath, err := extractBinary(tmpFile.Name())
	if err != nil {
		return fmt.Errorf("failed to extract binary: %w", err)
	}
	defer os.Remove(binaryPath)

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	if err := replaceBinary(binaryPath, execPath); err != nil {
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	return nil
}

func extractBinary(archivePath string) (string, error) {
	return extractFromTarGz(archivePath)
}

func extractFromTarGz(archivePath string) (string, error) {
	file, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return "", err
	}
	defer gzr.Close()

	return extractBinaryFromStream(gzr)
}

func extractBinaryFromStream(r io.Reader) (string, error) {
	tr := tar.NewReader(r)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		// Look for the astat binary
		if filepath.Base(header.Name) == "astat" {
			return saveBinary(tr)
		}
	}

	return "", fmt.Errorf("binary not found in archive")
}

func saveBinary(r io.Reader) (string, error) {
	tmpBinary, err := os.CreateTemp("", "astat-new-*")
	if err != nil {
		return "", err
	}
	tmpPath := tmpBinary.Name()

	if _, err := io.Copy(tmpBinary, r); err != nil {
		tmpBinary.Close()
		os.Remove(tmpPath)
		return "", err
	}
	tmpBinary.Close()

	// Make executable
	if err := os.Chmod(tmpPath, 0755); err != nil {
		os.Remove(tmpPath)
		return "", err
	}

	return tmpPath, nil
}

func replaceBinary(newBinary, targetPath string) error {
	// Create backup
	backupPath := targetPath + ".backup"
	if err := os.Rename(targetPath, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Copy new binary to target location
	if err := copyFile(newBinary, targetPath); err != nil {
		// Restore backup on failure
		os.Rename(backupPath, targetPath)
		return fmt.Errorf("failed to copy new binary: %w", err)
	}

	// Make executable
	if err := os.Chmod(targetPath, 0755); err != nil {
		// Restore backup on failure
		os.Remove(targetPath)
		os.Rename(backupPath, targetPath)
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Remove backup
	os.Remove(backupPath)

	return nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
