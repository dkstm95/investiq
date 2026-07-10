package app

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReleaseAssetNameUsesWindowsZip(t *testing.T) {
	if got := releaseAssetName("windows", "amd64"); got != "gacha-windows-amd64.zip" {
		t.Fatalf("unexpected Windows asset: %s", got)
	}
	if got := releaseAssetName("linux", "amd64"); got != "gacha-linux-amd64.tar.gz" {
		t.Fatalf("unexpected Linux asset: %s", got)
	}
}

func TestReleaseAssetURLForUsesTargetArchive(t *testing.T) {
	got := releaseAssetURLFor("v0.1.27", "windows", "arm64")
	want := "https://github.com/dkstm95/gacha/releases/download/v0.1.27/gacha-windows-arm64.zip"
	if got != want {
		t.Fatalf("unexpected URL:\n got: %s\nwant: %s", got, want)
	}
}

func TestReleaseChecksumsURL(t *testing.T) {
	got := releaseChecksumsURL("v0.1.37")
	want := "https://github.com/dkstm95/gacha/releases/download/v0.1.37/checksums.txt"
	if got != want {
		t.Fatalf("unexpected checksums URL:\n got: %s\nwant: %s", got, want)
	}
}

func TestCompareReleaseVersionsPreventsDowngrade(t *testing.T) {
	for _, testCase := range []struct {
		current string
		target  string
		want    int
	}{
		{current: "0.3.0", target: "v0.2.5", want: 1},
		{current: "v0.2.5", target: "0.2.5", want: 0},
		{current: "0.2.5", target: "0.2.6", want: -1},
	} {
		got, ok := compareReleaseVersions(testCase.current, testCase.target)
		if !ok || got != testCase.want {
			t.Fatalf("compareReleaseVersions(%q, %q) = %d, %v; want %d, true", testCase.current, testCase.target, got, ok, testCase.want)
		}
	}
}

func TestLatestReleaseTagUsesConfiguredEndpoint(t *testing.T) {
	oldURL := gitHubLatestReleaseURL
	gitHubLatestReleaseURL = "https://example.test/latest"
	app := New("test")
	app.env.HTTPClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != "https://example.test/latest" {
			t.Fatalf("unexpected URL: %s", req.URL.String())
		}
		return stringResponse(http.StatusOK, `{"tag_name":"v9.9.9"}`), nil
	})}
	t.Cleanup(func() { gitHubLatestReleaseURL = oldURL })

	got, err := app.latestReleaseTag()
	if err != nil {
		t.Fatal(err)
	}
	if got != "v9.9.9" {
		t.Fatalf("unexpected release tag: %s", got)
	}
}

func TestDownloadFileReportsHTTPFailures(t *testing.T) {
	app := New("test")
	app.env.HTTPClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return stringResponse(http.StatusNotFound, "missing"), nil
	})}

	err := app.downloadFile("https://example.test/missing", filepath.Join(t.TempDir(), "out"))
	if err == nil || !strings.Contains(err.Error(), "404") {
		t.Fatalf("expected HTTP failure, got %v", err)
	}
}

func TestDownloadFileWritesResponseBody(t *testing.T) {
	app := New("test")
	app.env.HTTPClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return stringResponse(http.StatusOK, "body"), nil
	})}
	path := filepath.Join(t.TempDir(), "out")
	if err := app.downloadFile("https://example.test/asset", path); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "body" {
		t.Fatalf("unexpected download body: %q", data)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func stringResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

func TestVerifyReleaseChecksum(t *testing.T) {
	dir := t.TempDir()
	archive := filepath.Join(dir, "gacha-linux-amd64.tar.gz")
	if err := os.WriteFile(archive, []byte("archive"), 0o600); err != nil {
		t.Fatal(err)
	}
	sum, err := sha256File(archive)
	if err != nil {
		t.Fatal(err)
	}
	checksums := filepath.Join(dir, "checksums.txt")
	if err := os.WriteFile(checksums, []byte(sum+"  gacha-linux-amd64.tar.gz\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := verifyReleaseChecksum(archive, checksums, "gacha-linux-amd64.tar.gz"); err != nil {
		t.Fatal(err)
	}
}

func TestVerifyReleaseChecksumRejectsMismatchAndMissingAsset(t *testing.T) {
	dir := t.TempDir()
	archive := filepath.Join(dir, "gacha-linux-amd64.tar.gz")
	if err := os.WriteFile(archive, []byte("archive"), 0o600); err != nil {
		t.Fatal(err)
	}
	checksums := filepath.Join(dir, "checksums.txt")
	if err := os.WriteFile(checksums, []byte("0  gacha-linux-amd64.tar.gz\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := verifyReleaseChecksum(archive, checksums, "gacha-linux-amd64.tar.gz"); err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("expected checksum mismatch, got %v", err)
	}
	if err := verifyReleaseChecksum(archive, checksums, "gacha-darwin-arm64.tar.gz"); err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected missing checksum, got %v", err)
	}
}

func TestExtractTarGzWritesGachaBinary(t *testing.T) {
	dir := t.TempDir()
	archive := filepath.Join(dir, "gacha.tar.gz")
	writeTarGz(t, archive, map[string]string{"gacha": "binary"})

	if err := extractTarGz(archive, dir); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "gacha"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "binary" {
		t.Fatalf("unexpected extracted binary: %q", data)
	}
}

func TestExtractTarGzRejectsArchiveWithoutGachaBinary(t *testing.T) {
	dir := t.TempDir()
	archive := filepath.Join(dir, "gacha.tar.gz")
	writeTarGz(t, archive, map[string]string{"README.md": "not a binary"})

	err := extractTarGz(archive, dir)
	if err == nil || !strings.Contains(err.Error(), "did not contain gacha binary") {
		t.Fatalf("expected missing binary error, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "gacha")); !os.IsNotExist(err) {
		t.Fatalf("unexpected gacha path after rejected archive: %v", err)
	}
}

func TestExtractTarGzIgnoresTraversalEntry(t *testing.T) {
	dir := t.TempDir()
	archive := filepath.Join(dir, "gacha.tar.gz")
	writeTarGz(t, archive, map[string]string{"../gacha": "escape"})

	err := extractTarGz(archive, dir)
	if err == nil || !strings.Contains(err.Error(), "did not contain gacha binary") {
		t.Fatalf("expected missing binary error, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(dir), "gacha")); !os.IsNotExist(err) {
		t.Fatalf("unexpected traversal output: %v", err)
	}
}

func TestReplaceExecutablePreservesOldFileUntilStagedCopySucceeds(t *testing.T) {
	dir := t.TempDir()
	destination := filepath.Join(dir, "gacha")
	if err := os.WriteFile(destination, []byte("old binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := replaceExecutable(filepath.Join(dir, "missing"), destination); err == nil {
		t.Fatal("expected staging failure")
	}
	data, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "old binary" {
		t.Fatalf("failed update changed existing executable: %q", data)
	}

	source := filepath.Join(dir, "new-gacha")
	if err := os.WriteFile(source, []byte("new binary"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := replaceExecutable(source, destination); err != nil {
		t.Fatal(err)
	}
	data, err = os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "new binary" {
		t.Fatalf("updated executable has unexpected content: %q", data)
	}
}

func TestSelfUpdateUnsupportedOnWindows(t *testing.T) {
	if selfUpdateSupported("windows") {
		t.Fatal("expected Windows self-update to be disabled")
	}
	if !selfUpdateSupported("linux") || !selfUpdateSupported("darwin") {
		t.Fatal("expected Unix self-update to stay enabled")
	}
	if !strings.Contains(windowsUpdateUnsupportedMessage(), "gacha-windows-amd64.zip") {
		t.Fatal("Windows update message should point users to Windows ZIP artifacts")
	}
}

func writeTarGz(t *testing.T, path string, files map[string]string) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()
	for name, content := range files {
		data := []byte(content)
		if err := tarWriter.WriteHeader(&tar.Header{
			Name: name,
			Mode: 0o755,
			Size: int64(len(data)),
		}); err != nil {
			t.Fatal(err)
		}
		if _, err := tarWriter.Write(data); err != nil {
			t.Fatal(err)
		}
	}
}
