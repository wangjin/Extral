package ffmpeg

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestResolvePaths_ReturnsNonEmpty(t *testing.T) {
	m := NewManager()
	ffmpegPath, ffprobePath := m.Paths()
	if ffmpegPath == "" {
		t.Error("ffmpeg path should not be empty")
	}
	if ffprobePath == "" {
		t.Error("ffprobe path should not be empty")
	}
}

func TestResolvePaths_PlatformExtension(t *testing.T) {
	m := NewManager()
	ffmpegPath, _ := m.Paths()
	if runtime.GOOS == "windows" {
		if filepath.Ext(ffmpegPath) != ".exe" {
			t.Error("Windows ffmpeg should have .exe extension")
		}
	} else {
		if filepath.Ext(ffmpegPath) == ".exe" {
			t.Error("Non-Windows ffmpeg should not have .exe extension")
		}
	}
}

func TestResolvePaths_RelativeToExecutable(t *testing.T) {
	m := NewManager()
	ffmpegPath, _ := m.Paths()

	if runtime.GOOS == "darwin" {
		if !containsPath(ffmpegPath, "Resources") {
			// May fall back to executable dir if no .app bundle
			execDir := filepath.Dir(m.execPath)
			if !containsPath(ffmpegPath, filepath.Base(execDir)) && ffmpegPath != filepath.Join(execDir, "ffmpeg") {
				t.Errorf("macOS ffmpeg path unexpected: %s", ffmpegPath)
			}
		}
	}
}

func containsPath(path, substr string) bool {
	p := filepath.Clean(path)
	for {
		if filepath.Base(p) == substr {
			return true
		}
		parent := filepath.Dir(p)
		if parent == p {
			break
		}
		p = parent
	}
	return false
}

func TestFFmpegExists(t *testing.T) {
	m := NewManager()
	_ = m.FFmpegExists()
}

func TestEnsureBinaries_CreatesTempFallback(t *testing.T) {
	m := NewManager()
	paths := m.FallbackPaths("/tmp/test-ffmpeg")
	if runtime.GOOS == "windows" {
		if paths[0] != "/tmp/test-ffmpeg/ffmpeg.exe" {
			t.Errorf("unexpected fallback path: %s", paths[0])
		}
	} else {
		if paths[0] != "/tmp/test-ffmpeg/ffmpeg" {
			t.Errorf("unexpected fallback path: %s", paths[0])
		}
	}
}

func TestParseProgress(t *testing.T) {
	tests := []struct {
		line        string
		wantFrame   int
		wantTime    float64
		wantOK      bool
	}{
		{"frame=  120 fps= 30 q=5.0 size=    1024kB time=00:00:04.00 bitrate= 209.7kbits/s speed=   1x", 120, 4.0, true},
		{"frame=    1 fps=0.0 q=0.0 size=       0kB time=00:00:00.00 bitrate=N/A speed=N/A", 1, 0, true},
		{"some other output", 0, 0, false},
	}
	for _, tt := range tests {
		frame, timeSec, ok := ParseProgress(tt.line)
		if ok != tt.wantOK {
			t.Errorf("ParseProgress(%q) ok = %v, want %v", tt.line, ok, tt.wantOK)
		}
		if frame != tt.wantFrame {
			t.Errorf("ParseProgress(%q) frame = %d, want %d", tt.line, frame, tt.wantFrame)
		}
		if tt.wantTime > 0 && (timeSec < tt.wantTime-0.5 || timeSec > tt.wantTime+0.5) {
			t.Errorf("ParseProgress(%q) timeSec = %f, want ~%f", tt.line, timeSec, tt.wantTime)
		}
	}
}
