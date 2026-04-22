package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

const rtkRepo = "rtk-ai/rtk"

// installRTK downloads and installs the rtk binary next to the maple binary.
// Returns the path to the installed binary, or an error.
// If rtk is already on PATH, returns the existing path immediately.
func installRTK() (string, error) {
	// already installed?
	if p, err := exec.LookPath("rtk"); err == nil {
		return p, nil
	}

	version, err := latestRTKVersion()
	if err != nil {
		return "", fmt.Errorf("could not resolve latest rtk version: %w", err)
	}

	triple, ext, err := rtkPlatformTriple()
	if err != nil {
		return "", err
	}

	archiveName := fmt.Sprintf("rtk-%s.%s", triple, ext)
	url := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", rtkRepo, version, archiveName)

	// download to a temp file
	tmp, err := os.CreateTemp("", "rtk-download-*")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmp.Name())

	resp, err := http.Get(url) //nolint:noctx
	if err != nil {
		return "", fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("download failed: HTTP %d for %s", resp.StatusCode, url)
	}
	if _, err := io.Copy(tmp, resp.Body); err != nil {
		return "", err
	}
	tmp.Close()

	// extract
	binData, err := extractRTKBinary(tmp.Name(), ext)
	if err != nil {
		return "", fmt.Errorf("extract failed: %w", err)
	}

	// install next to the maple binary
	dest := rtkInstallPath()
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(dest, binData, 0o755); err != nil {
		return "", fmt.Errorf("write failed: %w", err)
	}

	return dest, nil
}

// latestRTKVersion fetches the highest semver tag from the RTK releases API.
func latestRTKVersion() (string, error) {
	url := "https://api.github.com/repos/" + rtkRepo + "/releases?per_page=100"
	resp, err := http.Get(url) //nolint:noctx
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var releases []struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return "", err
	}
	var tags []string
	for _, r := range releases {
		if r.TagName != "" {
			tags = append(tags, r.TagName)
		}
	}
	if len(tags) == 0 {
		return "", fmt.Errorf("no releases found")
	}
	sort.Slice(tags, func(i, j int) bool {
		return compareSemver(tags[i], tags[j]) < 0
	})
	return tags[len(tags)-1], nil
}

// rtkPlatformTriple returns the release asset triple and archive extension for the current OS/arch.
func rtkPlatformTriple() (triple, ext string, err error) {
	os_ := runtime.GOOS
	arch := runtime.GOARCH
	switch os_ + "/" + arch {
	case "darwin/arm64":
		return "aarch64-apple-darwin", "tar.gz", nil
	case "darwin/amd64":
		return "x86_64-apple-darwin", "tar.gz", nil
	case "linux/arm64":
		return "aarch64-unknown-linux-gnu", "tar.gz", nil
	case "linux/amd64":
		return "x86_64-unknown-linux-gnu", "tar.gz", nil
	case "windows/amd64":
		return "x86_64-pc-windows-msvc", "zip", nil
	default:
		return "", "", fmt.Errorf("rtk: unsupported platform %s/%s", os_, arch)
	}
}

// rtkInstallPath returns where to install the rtk binary — same directory as maple.
func rtkInstallPath() string {
	exe, err := os.Executable()
	if err != nil {
		return filepath.Join(os.Getenv("HOME"), ".tools", "maple", "bin", rtkBinaryName())
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	return filepath.Join(filepath.Dir(exe), rtkBinaryName())
}

func rtkBinaryName() string {
	if runtime.GOOS == "windows" {
		return "rtk.exe"
	}
	return "rtk"
}

// extractRTKBinary finds and returns the raw bytes of the rtk binary inside the archive.
func extractRTKBinary(archivePath, ext string) ([]byte, error) {
	switch ext {
	case "tar.gz":
		return extractFromTarGz(archivePath)
	case "zip":
		return extractFromZip(archivePath)
	default:
		return nil, fmt.Errorf("unknown archive extension: %s", ext)
	}
}

func extractFromTarGz(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		base := filepath.Base(hdr.Name)
		if base == "rtk" || base == "rtk.exe" || isRTKBinaryName(hdr.Name) {
			return io.ReadAll(tr)
		}
	}
	return nil, fmt.Errorf("rtk binary not found in archive")
}

func extractFromZip(path string) ([]byte, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	for _, f := range r.File {
		base := filepath.Base(f.Name)
		if base == "rtk" || base == "rtk.exe" || isRTKBinaryName(f.Name) {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			return io.ReadAll(rc)
		}
	}
	return nil, fmt.Errorf("rtk binary not found in archive")
}

// isRTKBinaryName matches binaries named after platform triples (e.g. rtk-x86_64-unknown-linux-gnu).
func isRTKBinaryName(name string) bool {
	base := strings.ToLower(filepath.Base(name))
	return strings.HasPrefix(base, "rtk") && !strings.Contains(base, ".")
}

// compareSemver compares two version strings lexicographically after stripping "v" prefix.
// Returns negative if a < b, zero if equal, positive if a > b.
func compareSemver(a, b string) int {
	normalize := func(s string) []int {
		s = strings.TrimPrefix(s, "v")
		parts := strings.Split(s, ".")
		nums := make([]int, 3)
		for i, p := range parts {
			if i >= 3 {
				break
			}
			n := 0
			for _, c := range p {
				if c >= '0' && c <= '9' {
					n = n*10 + int(c-'0')
				} else {
					break
				}
			}
			nums[i] = n
		}
		return nums
	}
	av, bv := normalize(a), normalize(b)
	for i := range av {
		if av[i] != bv[i] {
			if av[i] < bv[i] {
				return -1
			}
			return 1
		}
	}
	return 0
}
