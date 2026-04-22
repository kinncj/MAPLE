package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"os"
	"runtime"
	"strings"
	"testing"
)

// ─── compareSemver ────────────────────────────────────────────────────────────

func TestCompareSemver(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"v1.0.0", "v1.0.0", 0},
		{"v1.0.0", "v1.0.1", -1},
		{"v1.0.1", "v1.0.0", 1},
		{"v2.0.0", "v1.9.9", 1},
		{"v0.9.9", "v1.0.0", -1},
		{"v4.9.2", "v4.10.0", -1},  // 9 < 10 numerically
		{"v4.10.0", "v4.9.2", 1},
		{"1.0.0", "v1.0.0", 0},     // no "v" prefix
	}
	for _, c := range cases {
		got := compareSemver(c.a, c.b)
		// normalise to -1/0/1
		if got < 0 {
			got = -1
		} else if got > 0 {
			got = 1
		}
		if got != c.want {
			t.Errorf("compareSemver(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

// ─── rtkBinaryName ────────────────────────────────────────────────────────────

func TestRTKBinaryName(t *testing.T) {
	name := rtkBinaryName()
	if runtime.GOOS == "windows" {
		if name != "rtk.exe" {
			t.Errorf("expected rtk.exe on windows, got %q", name)
		}
	} else {
		if name != "rtk" {
			t.Errorf("expected rtk on non-windows, got %q", name)
		}
	}
}

// ─── rtkPlatformTriple ───────────────────────────────────────────────────────

func TestRTKPlatformTriple(t *testing.T) {
	triple, ext, err := rtkPlatformTriple()
	if err != nil {
		// only an error on unsupported platforms — ok to skip
		t.Skipf("unsupported platform: %v", err)
	}
	if triple == "" {
		t.Error("triple should not be empty")
	}
	if ext != "tar.gz" && ext != "zip" {
		t.Errorf("unexpected ext %q", ext)
	}
	if runtime.GOOS == "windows" && ext != "zip" {
		t.Error("expected zip on windows")
	}
	if runtime.GOOS != "windows" && ext != "tar.gz" {
		t.Error("expected tar.gz on non-windows")
	}
}

// ─── isRTKBinaryName ─────────────────────────────────────────────────────────

func TestIsRTKBinaryName(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"rtk", true},
		{"rtk.exe", false},                          // has a dot
		{"rtk-x86_64-unknown-linux-gnu", true},
		{"rtk-aarch64-apple-darwin", true},
		{"something-else", false},
		{"rtk-install.sh", false},                   // has a dot
		{"rtk-v1.0.0", false},                       // has a dot
		{"subdir/rtk-x86_64-unknown-linux-gnu", true},
	}
	for _, c := range cases {
		got := isRTKBinaryName(c.path)
		if got != c.want {
			t.Errorf("isRTKBinaryName(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

// ─── extractFromTarGz ────────────────────────────────────────────────────────

func TestExtractFromTarGz_RootLevel(t *testing.T) {
	// build a tar.gz with "rtk" at the root
	data := makeTarGz(t, map[string][]byte{
		"rtk": []byte("fake-rtk-binary"),
	})
	f := writeTempFile(t, data, "*.tar.gz")
	got, err := extractFromTarGz(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != "fake-rtk-binary" {
		t.Errorf("got %q, want %q", got, "fake-rtk-binary")
	}
}

func TestExtractFromTarGz_Subdirectory(t *testing.T) {
	// build a tar.gz with "rtk" inside a platform-named subdirectory
	data := makeTarGz(t, map[string][]byte{
		"rtk-x86_64-unknown-linux-gnu/rtk": []byte("fake-rtk-binary"),
	})
	f := writeTempFile(t, data, "*.tar.gz")
	got, err := extractFromTarGz(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != "fake-rtk-binary" {
		t.Errorf("got %q, want %q", got, "fake-rtk-binary")
	}
}

func TestExtractFromTarGz_TripleBinaryName(t *testing.T) {
	// some releases ship the binary named after the triple with no extension
	data := makeTarGz(t, map[string][]byte{
		"rtk-aarch64-apple-darwin": []byte("fake-rtk-binary"),
	})
	f := writeTempFile(t, data, "*.tar.gz")
	got, err := extractFromTarGz(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != "fake-rtk-binary" {
		t.Errorf("got %q, want %q", got, "fake-rtk-binary")
	}
}

func TestExtractFromTarGz_NotFound(t *testing.T) {
	data := makeTarGz(t, map[string][]byte{
		"README.md": []byte("not the binary"),
	})
	f := writeTempFile(t, data, "*.tar.gz")
	_, err := extractFromTarGz(f)
	if err == nil {
		t.Error("expected error when binary not found")
	}
}

// ─── extractFromZip ──────────────────────────────────────────────────────────

func TestExtractFromZip_RootLevel(t *testing.T) {
	data := makeZip(t, map[string][]byte{
		"rtk.exe": []byte("fake-rtk-binary"),
	})
	f := writeTempFile(t, data, "*.zip")
	got, err := extractFromZip(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != "fake-rtk-binary" {
		t.Errorf("got %q, want %q", got, "fake-rtk-binary")
	}
}

func TestExtractFromZip_Subdirectory(t *testing.T) {
	data := makeZip(t, map[string][]byte{
		"rtk-x86_64-pc-windows-msvc/rtk.exe": []byte("fake-rtk-binary"),
	})
	f := writeTempFile(t, data, "*.zip")
	got, err := extractFromZip(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != "fake-rtk-binary" {
		t.Errorf("got %q, want %q", got, "fake-rtk-binary")
	}
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func makeTarGz(t *testing.T, files map[string][]byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	for name, content := range files {
		if err := tw.WriteHeader(&tar.Header{
			Name: name,
			Mode: 0o755,
			Size: int64(len(content)),
		}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write(content); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func makeZip(t *testing.T, files map[string][]byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write(content); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func writeTempFile(t *testing.T, data []byte, pattern string) string {
	t.Helper()
	f, err := os.CreateTemp("", pattern)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Remove(f.Name()) })
	if _, err := f.Write(data); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

// ─── integration: latestRTKVersion (network, skipped in CI if no network) ────

func TestLatestRTKVersion_Format(t *testing.T) {
	if os.Getenv("SKIP_NETWORK_TESTS") != "" {
		t.Skip("network tests disabled")
	}
	ver, err := latestRTKVersion()
	if err != nil {
		t.Skipf("network unavailable: %v", err)
	}
	if !strings.HasPrefix(ver, "v") && !strings.Contains(ver, ".") {
		t.Errorf("unexpected version format %q", ver)
	}
}
