package app

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type GitHubRelease struct {
	TagName string `json:"tag_name"`
}

var (
	gitHubLatestReleaseURL = "https://api.github.com/repos/dkstm95/gacha/releases/latest"
	gitHubReleaseBaseURL   = "https://github.com/dkstm95/gacha/releases/download"
)

func (a *App) updateSelf() error {
	if !selfUpdateSupported(a.env.GOOS) {
		fmt.Println(windowsUpdateUnsupportedMessage())
		return nil
	}

	latest, err := a.latestReleaseTag()
	if err != nil {
		return err
	}
	current := normalizeVersion(a.version)
	target := normalizeVersion(latest)
	if comparison, ok := compareReleaseVersions(current, target); (ok && comparison >= 0) || current == target {
		fmt.Printf("gacha is already up to date (%s).\n", a.version)
		return nil
	}

	exe, err := os.Executable()
	if err != nil {
		return err
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return err
	}

	fmt.Printf("Updating gacha %s -> %s\n", a.version, target)
	tmpDir, err := os.MkdirTemp("", "gacha-update-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	assetName := releaseAssetName(a.env.GOOS, a.env.GOARCH)
	archivePath := filepath.Join(tmpDir, assetName)
	url := releaseAssetURLFor(latest, a.env.GOOS, a.env.GOARCH)
	if err := a.downloadFile(url, archivePath); err != nil {
		return fmt.Errorf("could not download update archive: %w\nManual install: https://github.com/dkstm95/gacha/releases/latest", err)
	}
	checksumsPath := filepath.Join(tmpDir, "checksums.txt")
	if err := a.downloadFile(releaseChecksumsURL(latest), checksumsPath); err != nil {
		return fmt.Errorf("could not download release checksums: %w\nManual install: https://github.com/dkstm95/gacha/releases/latest", err)
	}
	if err := verifyReleaseChecksum(archivePath, checksumsPath, assetName); err != nil {
		return fmt.Errorf("could not verify update download: %w\nManual install: https://github.com/dkstm95/gacha/releases/latest", err)
	}
	if err := extractTarGz(archivePath, tmpDir); err != nil {
		return fmt.Errorf("could not unpack update archive: %w", err)
	}
	newBinary := filepath.Join(tmpDir, "gacha")
	if err := os.Chmod(newBinary, 0o755); err != nil {
		return fmt.Errorf("could not prepare update binary: %w", err)
	}

	if err := replaceExecutable(newBinary, exe); err != nil {
		return fmt.Errorf("could not replace %s: %w\nTry installing manually from https://github.com/dkstm95/gacha/releases/latest", exe, err)
	}

	fmt.Printf("Updated %s to %s.\n", exe, target)
	return nil
}

func (a *App) latestReleaseTag() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, gitHubLatestReleaseURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "gacha/"+a.version)
	resp, err := a.env.httpClient().Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("GitHub release check failed: %s", resp.Status)
	}
	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}
	if release.TagName == "" {
		return "", fmt.Errorf("GitHub latest release did not include a tag")
	}
	return release.TagName, nil
}

func normalizeVersion(value string) string {
	return strings.TrimPrefix(strings.TrimSpace(value), "v")
}

func compareReleaseVersions(current string, target string) (int, bool) {
	currentParts, ok := numericVersionParts(current)
	if !ok {
		return 0, false
	}
	targetParts, ok := numericVersionParts(target)
	if !ok {
		return 0, false
	}
	for index := range currentParts {
		if currentParts[index] > targetParts[index] {
			return 1, true
		}
		if currentParts[index] < targetParts[index] {
			return -1, true
		}
	}
	return 0, true
}

func numericVersionParts(value string) ([3]int, bool) {
	var result [3]int
	value = strings.SplitN(normalizeVersion(value), "+", 2)[0]
	value = strings.SplitN(value, "-", 2)[0]
	parts := strings.Split(value, ".")
	if len(parts) == 0 || len(parts) > len(result) {
		return result, false
	}
	for index, part := range parts {
		number, err := strconv.Atoi(part)
		if err != nil || number < 0 {
			return result, false
		}
		result[index] = number
	}
	return result, true
}

func releaseAssetURL(tag string) string {
	return releaseAssetURLFor(tag, runtime.GOOS, runtime.GOARCH)
}

func releaseAssetURLFor(tag string, goos string, goarch string) string {
	return fmt.Sprintf("%s/%s/%s", strings.TrimRight(gitHubReleaseBaseURL, "/"), tag, releaseAssetName(goos, goarch))
}

func releaseChecksumsURL(tag string) string {
	return fmt.Sprintf("%s/%s/checksums.txt", strings.TrimRight(gitHubReleaseBaseURL, "/"), tag)
}

func releaseAssetName(goos string, goarch string) string {
	extension := ".tar.gz"
	if goos == "windows" {
		extension = ".zip"
	}
	return "gacha-" + targetTripleFor(goos, goarch) + extension
}

func verifyReleaseChecksum(archivePath string, checksumsPath string, assetName string) error {
	expected, err := checksumForAsset(checksumsPath, assetName)
	if err != nil {
		return err
	}
	actual, err := sha256File(archivePath)
	if err != nil {
		return err
	}
	if actual != expected {
		return fmt.Errorf("checksum mismatch for %s", assetName)
	}
	return nil
}

func checksumForAsset(checksumsPath string, assetName string) (string, error) {
	data, err := os.ReadFile(checksumsPath)
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[1] == assetName {
			return fields[0], nil
		}
	}
	return "", fmt.Errorf("checksum for %s not found", assetName)
}

func sha256File(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func selfUpdateSupported(goos string) bool {
	return goos != "windows"
}

func windowsUpdateUnsupportedMessage() string {
	return strings.Join([]string{
		"Windows self-update is not supported yet.",
		"Download the latest gacha-windows-amd64.zip or gacha-windows-arm64.zip from:",
		"https://github.com/dkstm95/gacha/releases/latest",
		"Then replace gacha.exe in your PATH and open a new terminal.",
	}, "\n")
}

func (a *App) downloadFile(url string, destination string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "gacha/"+a.version)
	resp, err := a.env.httpClient().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("download failed: %s", resp.Status)
	}
	out, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = out.ReadFrom(resp.Body)
	return err
}

func extractTarGz(archivePath string, destinationDir string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			return fmt.Errorf("release archive did not contain gacha binary")
		}
		if err != nil {
			return err
		}
		if header.Typeflag != tar.TypeReg {
			continue
		}
		name := filepath.Clean(header.Name)
		if name != "gacha" {
			continue
		}
		target := filepath.Join(destinationDir, "gacha")
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(out, tarReader)
		closeErr := out.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
		return nil
	}
}

func replaceExecutable(source string, destination string) error {
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()

	staged, err := os.CreateTemp(filepath.Dir(destination), ".gacha-update-*")
	if err != nil {
		return err
	}
	stagedPath := staged.Name()
	defer os.Remove(stagedPath)

	if _, err := io.Copy(staged, in); err != nil {
		_ = staged.Close()
		return err
	}
	if err := staged.Chmod(0o755); err != nil {
		_ = staged.Close()
		return err
	}
	if err := staged.Sync(); err != nil {
		_ = staged.Close()
		return err
	}
	if err := staged.Close(); err != nil {
		return err
	}
	return os.Rename(stagedPath, destination)
}

func targetTriple() string {
	return targetTripleFor(runtime.GOOS, runtime.GOARCH)
}

func targetTripleFor(goos string, goarch string) string {
	return goos + "-" + goarch
}
