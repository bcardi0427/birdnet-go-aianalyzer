package ffmpeg

import (
	"os"
	"path/filepath"
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
		`C:\ffmpeg\bin\ffmpeg.exe`,
		`C:\Program Files\FFmpeg\bin\ffmpeg.exe`,
		`C:\Program Files (x86)\FFmpeg\bin\ffmpeg.exe`,
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
	if fi, err := os.Stat(path); err == nil {
		if fi.Mode().IsRegular() {
			return true
		}
		// If path points to a file without execution bit, still consider as executable on Windows
	}
	// Try to resolve using filepath.Clean in case of quotes or spaces
	if abs, err := filepath.Abs(path); err == nil {
		if fi, err := os.Stat(abs); err == nil {
			return fi.Mode().IsRegular()
		}
	}
	return false
}
