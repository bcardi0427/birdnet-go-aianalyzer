package ffmpeg

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsExe(t *testing.T) {
	// Empty path
	assert.False(t, isExe(""))

	// Path traversal
	assert.False(t, isExe("../../foo/bar"))

	// Non-existent path
	assert.False(t, isExe("nonexistent_file_xyz.exe"))

	// Create a temp file to test valid executable checks
	tempDir, err := os.MkdirTemp("", "ffmpeg-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	tempFile := filepath.Join(tempDir, "dummy_exe.exe")
	err = os.WriteFile(tempFile, []byte("dummy exe content"), 0o755)
	if err != nil {
		t.Fatal(err)
	}

	assert.True(t, isExe(tempFile))
}

func TestGetFFmpegPath(t *testing.T) {
	// Mock BIRDNET_FFMPEG_PATH env var
	oldEnv := os.Getenv("BIRDNET_FFMPEG_PATH")
	defer os.Setenv("BIRDNET_FFMPEG_PATH", oldEnv)

	// Set to nonexistent path - should fall back to candidates
	os.Setenv("BIRDNET_FFMPEG_PATH", "nonexistent_ffmpeg_binary_xyz.exe")
	
	tempDir, err := os.MkdirTemp("", "ffmpeg-env-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	mockExe := filepath.Join(tempDir, "ffmpeg.exe")
	err = os.WriteFile(mockExe, []byte("ffmpeg"), 0o755)
	if err != nil {
		t.Fatal(err)
	}

	os.Setenv("BIRDNET_FFMPEG_PATH", mockExe)
	path := GetFFmpegPath()
	assert.Equal(t, mockExe, path)
}
