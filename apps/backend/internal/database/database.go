package database

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
)

const applicationName = "lwn-simulator"

//go:embed defaults/*.json
var defaults embed.FS

func DataDir() (string, error) {
	if os.Getenv("LWN_ENV") == "development" {
		return filepath.Abs("./dev-db")
	}

	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("APPDATA environment variable is not set")
		}

		return filepath.Join(appData, applicationName), nil

	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("get user home directory: %w", err)
		}

		return filepath.Join(
			home,
			"Library",
			"Application Support",
			applicationName,
		), nil

	default:
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("get user home directory: %w", err)
		}

		return filepath.Join(
			home,
			".local",
			"share",
			applicationName,
		), nil
	}
}

func Initialize() (string, error) {
	dataDir, err := DataDir()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return "", fmt.Errorf("create data directory: %w", err)
	}

	entries, err := fs.ReadDir(defaults, "defaults")
	if err != nil {
		return "", fmt.Errorf("read embedded defaults: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		if err := initializeFile(dataDir, entry.Name()); err != nil {
			return "", err
		}
	}

	return dataDir, nil
}

func initializeFile(dataDir string, name string) error {
	path := filepath.Join(dataDir, name)

	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("check database file %s: %w", name, err)
	}

	data, err := defaults.ReadFile(filepath.Join("defaults", name))
	if err != nil {
		return fmt.Errorf("read embedded default %s: %w", name, err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("create database file %s: %w", name, err)
	}

	return nil
}
