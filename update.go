package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	githubRepo   = "helloWorld44-89/tengo"
	maxBinaryMB  = 100 * 1024 * 1024 // 100 MB download cap
)

// platformAsset returns the release asset filename for the current OS/arch.
func platformAsset() (string, bool) {
	switch runtime.GOOS {
	case "linux":
		switch runtime.GOARCH {
		case "amd64":
			return "tengo-linux-amd64", true
		case "arm64":
			return "tengo-linux-arm64", true
		}
	case "darwin":
		switch runtime.GOARCH {
		case "amd64":
			return "tengo-darwin-amd64", true
		case "arm64":
			return "tengo-darwin-arm64", true
		}
	case "windows":
		if runtime.GOARCH == "amd64" {
			return "tengo-windows-amd64.exe", true
		}
	}
	return "", false
}

type ghRelease struct {
	TagName string `json:"tag_name"`
}

func fetchLatestTag() (string, error) {
	url := "https://api.github.com/repos/" + githubRepo + "/releases/latest"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "tengo/"+version)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned HTTP %d", resp.StatusCode)
	}

	var rel ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return "", err
	}
	if rel.TagName == "" {
		return "", fmt.Errorf("no releases found")
	}
	return rel.TagName, nil
}

func runUpdate(quiet bool) int {
	asset, ok := platformAsset()
	if !ok {
		fmt.Fprintf(os.Stderr, "No prebuilt binary for %s/%s — build from source instead.\n",
			runtime.GOOS, runtime.GOARCH)
		return 1
	}

	if !quiet {
		fmt.Printf("Current version: %s\nChecking for updates...\n", version)
	}

	latest, err := fetchLatestTag()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Update check failed: %v\n", err)
		return 2
	}

	latestClean := strings.TrimPrefix(latest, "v")
	currentClean := strings.TrimPrefix(version, "v")

	if latestClean == currentClean {
		if !quiet {
			fmt.Printf("Already up to date (%s).\n", version)
		}
		return 0
	}

	if !quiet {
		fmt.Printf("Updating %s → %s...\n", version, latest)
	}

	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot determine executable path: %v\n", err)
		return 2
	}

	// Preserve existing permissions.
	info, err := os.Stat(exe)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot stat current binary: %v\n", err)
		return 2
	}

	downloadURL := "https://github.com/" + githubRepo + "/releases/latest/download/" + asset
	resp, err := http.Get(downloadURL) //nolint:noctx
	if err != nil {
		fmt.Fprintf(os.Stderr, "Download failed: %v\n", err)
		return 2
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Download failed: HTTP %d\n", resp.StatusCode)
		return 2
	}

	// Write to a temp file next to the current binary so rename is atomic.
	tmp, err := os.CreateTemp(filepath.Dir(exe), ".tengo-update-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot create temp file: %v\n", err)
		return 2
	}
	tmpName := tmp.Name()

	limited := io.LimitReader(resp.Body, maxBinaryMB+1)
	n, err := io.Copy(tmp, limited)
	tmp.Close()
	if err != nil {
		os.Remove(tmpName)
		fmt.Fprintf(os.Stderr, "Download failed: %v\n", err)
		return 2
	}
	if n > maxBinaryMB {
		os.Remove(tmpName)
		fmt.Fprintf(os.Stderr, "Downloaded binary is too large (> 100 MB); aborting.\n")
		return 2
	}

	if err := os.Chmod(tmpName, info.Mode()); err != nil {
		os.Remove(tmpName)
		fmt.Fprintf(os.Stderr, "Cannot set permissions: %v\n", err)
		return 2
	}

	if err := os.Rename(tmpName, exe); err != nil {
		os.Remove(tmpName)
		fmt.Fprintf(os.Stderr, "Cannot replace binary: %v\n", err)
		if runtime.GOOS == "windows" {
			fmt.Fprintln(os.Stderr, "On Windows, try running as Administrator.")
		} else {
			fmt.Fprintln(os.Stderr, "Try running with sudo.")
		}
		return 2
	}

	if !quiet {
		fmt.Printf("Updated to %s. Restart tengo to use the new version.\n", latest)
	}
	return 0
}
