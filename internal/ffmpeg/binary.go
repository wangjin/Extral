package ffmpeg

import (
	"fmt"
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
		ffmpeg = filepath.Join(dir, "ffmpeg.exe")
		ffprobe = filepath.Join(dir, "ffprobe.exe")
	} else {
		ffmpeg = filepath.Join(dir, "ffmpeg")
		ffprobe = filepath.Join(dir, "ffprobe")
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
