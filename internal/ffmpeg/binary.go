package ffmpeg

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

type Manager struct {
	execPath string
}

func NewManager() *Manager {
	execPath, _ := os.Executable()
	return &Manager{execPath: execPath}
}

func embeddedCacheDir() string {
	dir, err := os.UserCacheDir()
	if err != nil {
		dir = os.TempDir()
	}
	return filepath.Join(dir, "Extral", "ffmpeg")
}

func ExtractEmbedded() {
	names, err := fs.ReadDir(EmbeddedBinaries, "bin")
	if err != nil {
		return
	}

	cacheDir := embeddedCacheDir()
	os.MkdirAll(cacheDir, 0755)

	for _, entry := range names {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}

		data, err := EmbeddedBinaries.ReadFile("bin/" + name)
		if err != nil {
			continue
		}

		dest := filepath.Join(cacheDir, name)
		if existing, err := os.Stat(dest); err == nil && int(existing.Size()) == len(data) {
			continue
		}
		os.WriteFile(dest, data, 0755)
	}
}

func (m *Manager) Paths() (ffmpeg string, ffprobe string) {
	dir := filepath.Dir(m.execPath)

	if runtime.GOOS == "darwin" {
		resourcesDir := filepath.Join(filepath.Dir(dir), "Resources", "ffmpeg")
		ffmpeg = filepath.Join(resourcesDir, "ffmpeg")
		ffprobe = filepath.Join(resourcesDir, "ffprobe")
		if _, err := os.Stat(ffmpeg); os.IsNotExist(err) {
			ffmpeg = filepath.Join(dir, "ffmpeg")
			ffprobe = filepath.Join(dir, "ffprobe")
		}
	} else if runtime.GOOS == "windows" {
		// Prefer embedded binaries extracted to cache
		cacheDir := embeddedCacheDir()
		cacheFFmpeg := filepath.Join(cacheDir, "ffmpeg.exe")
		cacheFFprobe := filepath.Join(cacheDir, "ffprobe.exe")
		if _, err := os.Stat(cacheFFmpeg); err == nil {
			ffmpeg = cacheFFmpeg
			ffprobe = cacheFFprobe
		} else {
			ffmpeg = filepath.Join(dir, "ffmpeg.exe")
			ffprobe = filepath.Join(dir, "ffprobe.exe")
		}
	} else {
		ffmpeg = filepath.Join(dir, "ffmpeg")
		ffprobe = filepath.Join(dir, "ffprobe")
	}

	// Fallback: look in system PATH
	if _, err := os.Stat(ffmpeg); os.IsNotExist(err) {
		if p, err := exec.LookPath("ffmpeg"); err == nil {
			ffmpeg = p
		}
		if p, err := exec.LookPath("ffprobe"); err == nil {
			ffprobe = p
		}
	}
	return
}

func (m *Manager) FFmpegExists() bool {
	ffmpeg, _ := m.Paths()
	_, err := os.Stat(ffmpeg)
	return err == nil
}

func (m *Manager) FallbackPaths(baseDir string) [2]string {
	if runtime.GOOS == "windows" {
		return [2]string{
			filepath.Join(baseDir, "ffmpeg.exe"),
			filepath.Join(baseDir, "ffprobe.exe"),
		}
	}
	return [2]string{
		filepath.Join(baseDir, "ffmpeg"),
		filepath.Join(baseDir, "ffprobe"),
	}
}

func (m *Manager) RunFFprobe(args ...string) ([]byte, error) {
	_, ffprobe := m.Paths()
	cmd := exec.Command(ffprobe, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ffprobe failed: %w, output: %s", err, string(out))
	}
	return out, nil
}

func (m *Manager) RunFFmpeg(args ...string) *exec.Cmd {
	ffmpeg, _ := m.Paths()
	cmd := exec.Command(ffmpeg, args...)
	return cmd
}

var (
	frameRe = regexp.MustCompile(`frame=\s*(\d+)`)
	timeRe  = regexp.MustCompile(`time=(\d+:\d+:\d+\.\d+)`)
)

func ParseProgress(line string) (frame int, timeSec float64, ok bool) {
	if !strings.Contains(line, "frame=") {
		return 0, 0, false
	}

	if m := frameRe.FindStringSubmatch(line); len(m) > 1 {
		fmt.Sscanf(m[1], "%d", &frame)
	}
	if m := timeRe.FindStringSubmatch(line); len(m) > 1 {
		timeSec = parseTimeToSeconds(m[1])
	}
	if frame > 0 || timeSec > 0 {
		ok = true
	}
	return
}

func parseTimeToSeconds(timeStr string) float64 {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 3 {
		return 0
	}
	var h, m, s float64
	fmt.Sscanf(parts[0], "%f", &h)
	fmt.Sscanf(parts[1], "%f", &m)
	fmt.Sscanf(parts[2], "%f", &s)
	return h*3600 + m*60 + s
}
