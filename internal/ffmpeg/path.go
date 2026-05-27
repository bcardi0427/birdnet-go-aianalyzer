package ffmpeg

import (
	"os"
	"path/filepath"
)

// Common Windows installation paths for FFmpeg.
const (
	winInstallPath1 = `C:\ffmpeg\bin\ffmpeg.exe`
	winInstallPath2 = `C:\Program Files\FFmpeg\bin\ffmpeg.exe`
	winInstallPath3 = `C:\Program Files (x86)\FFmpeg\bin\ffmpeg.exe`
)

// GetFFmpegPath returns the configured FFmpeg binary path.
// It checks an environment variable first, then common install locations.
func GetFFmpegPath() string {
	if p := os.Getenv("BIRDNET_FFMPEG_PATH"); p != "" {
		if isExe(p) {
			return p
		}
	}
	// Common Windows install locations
	candidates := []string{
		winInstallPath1,
		winInstallPath2,
		winInstallPath3,
	}
	for _, c := range candidates {
		if isExe(c) {
			return c
		}
	}
	// If not found, return empty and let caller handle fallback.
	return ""
}

func isExe(path string) bool {
	if path == "" {
		return false
	}

	// Clean the path to resolve any traversal elements
	cleaned := filepath.Clean(path)

	// Validate against path traversal for relative paths
	if !filepath.IsAbs(cleaned) && !filepath.IsLocal(cleaned) {
		return false
	}

	if fi, err := os.Stat(cleaned); err == nil {
		if fi.Mode().IsRegular() {
			return true
		}
	}

	// Try to resolve using absolute path in case of relative path checks
	if abs, err := filepath.Abs(cleaned); err == nil {
		if fi, err := os.Stat(abs); err == nil {
			return fi.Mode().IsRegular()
		}
	}
	return false
}
