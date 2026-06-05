package recorder

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"

	"video-extractor/internal/ffmpeg"
)

var (
	recSizeRe = regexp.MustCompile(`size=\s*(\d+)kB`)
	recFpsRe  = regexp.MustCompile(`fps=\s*([\d.]+)`)
	recDropRe = regexp.MustCompile(`drop=\s*(\d+)`)
)

// Session manages a recording session lifecycle.
type Session struct {
	mu      sync.Mutex
	ffmpeg  *ffmpeg.Manager
	cmd     *exec.Cmd
	cancel  context.CancelFunc
	config  RecordingConfig
	camera  Camera
	started time.Time
	paused  bool
}

// NewSession creates a new Session instance.
func NewSession(ffmpegMgr *ffmpeg.Manager) *Session {
	return &Session{ffmpeg: ffmpegMgr}
}

// Start begins recording with the given configuration.
func (s *Session) Start(config RecordingConfig, camera Camera) error {
	s.Stop()

	// Ensure output directory exists
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	outputPath := config.OutputDir + "/" + config.FileName

	ctx, cancel := context.WithCancel(context.Background())
	args := BuildRecordingArgs(config, camera, outputPath)
	cmd := s.ffmpeg.RunFFmpeg(args...)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return fmt.Errorf("获取 stderr 失败: %w", err)
	}

	if err := cmd.Start(); err != nil {
		cancel()
		return fmt.Errorf("启动 FFmpeg 录制失败: %w", err)
	}

	s.mu.Lock()
	s.cmd = cmd
	s.cancel = cancel
	s.config = config
	s.camera = camera
	s.started = time.Now()
	s.paused = false
	s.mu.Unlock()

	ffmpegPath, _ := s.ffmpeg.Paths()
	application.Get().Event.Emit("recorder:log",
		fmt.Sprintf("[命令] %s %s", ffmpegPath, strings.Join(args, " ")))

	go s.readStatus(ctx, stderr)

	return nil
}

// Stop terminates the recording session gracefully.
func (s *Session) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	if s.cmd != nil && s.cmd.Process != nil {
		// Send SIGINT for graceful shutdown (ensures file header is written)
		if runtime.GOOS != "windows" {
			s.cmd.Process.Signal(syscall.SIGINT)
			done := make(chan error, 1)
			go func() {
				done <- s.cmd.Wait()
			}()
			select {
			case <-done:
			case <-time.After(5 * time.Second):
				s.cmd.Process.Kill()
				s.cmd.Wait()
			}
		} else {
			s.cmd.Process.Kill()
			s.cmd.Wait()
		}
	}
	s.cmd = nil
	s.paused = false
}

// Pause suspends the recording (macOS/Linux only).
func (s *Session) Pause() error {
	if runtime.GOOS == "windows" {
		return fmt.Errorf("Windows 不支持暂停录制")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cmd == nil || s.cmd.Process == nil {
		return fmt.Errorf("没有正在进行的录制")
	}
	if s.paused {
		return nil
	}
	if err := s.cmd.Process.Signal(syscall.SIGTSTP); err != nil {
		return fmt.Errorf("暂停失败: %w", err)
	}
	s.paused = true
	return nil
}

// Resume resumes a paused recording (macOS/Linux only).
func (s *Session) Resume() error {
	if runtime.GOOS == "windows" {
		return fmt.Errorf("Windows 不支持暂停录制")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cmd == nil || s.cmd.Process == nil {
		return fmt.Errorf("没有正在进行的录制")
	}
	if !s.paused {
		return nil
	}
	if err := s.cmd.Process.Signal(syscall.SIGCONT); err != nil {
		return fmt.Errorf("恢复失败: %w", err)
	}
	s.paused = false
	return nil
}

// IsRunning returns whether a recording is in progress.
func (s *Session) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cmd != nil && s.cmd.Process != nil
}

// IsPaused returns whether the recording is paused.
func (s *Session) IsPaused() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.paused
}

// readStatus reads FFmpeg stderr and emits status events.
func (s *Session) readStatus(ctx context.Context, stderr io.Reader) {
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		line := scanner.Text()
		application.Get().Event.Emit("recorder:log", line)
		s.emitStatusFromLine(line)
	}
	if err := scanner.Err(); err != nil {
		application.Get().Event.Emit("recorder:log",
			fmt.Sprintf("[stderr 读取错误] %v", err))
	}

	// FFmpeg exited
	s.mu.Lock()
	waitErr := s.cmd.Wait()
	s.mu.Unlock()

	select {
	case <-ctx.Done():
		return
	default:
	}

	state := "idle"
	errMsg := ""
	if waitErr != nil {
		errMsg = fmt.Sprintf("FFmpeg 退出异常: %v", waitErr)
		application.Get().Event.Emit("recorder:error",
			map[string]string{"error": errMsg})
	}

	application.Get().Event.Emit("recorder:status", RecordingStatus{
		State:    state,
		Duration: int64(time.Since(s.started).Seconds()),
		Error:    errMsg,
	})
	if errMsg == "" {
		application.Get().Event.Emit("recorder:log", "[完成]")
	}
}

// emitStatusFromLine parses an FFmpeg progress line and emits a status event.
func (s *Session) emitStatusFromLine(line string) {
	if !strings.Contains(line, "frame=") {
		return
	}

	_, _, ok := ffmpeg.ParseProgress(line)
	if !ok {
		return
	}

	fps := parseFloat(line, recFpsRe)
	sizeKB := parseFloat(line, recSizeRe)
	dropped := parseInt(line, recDropRe)

	status := RecordingStatus{
		State:    "recording",
		Duration: int64(time.Since(s.started).Seconds()),
		FPS:      fps,
	}
	s.mu.Lock()
	if s.paused {
		status.State = "paused"
	}
	s.mu.Unlock()

	if sizeKB > 0 {
		status.FileSize = int64(sizeKB * 1024)
	}
	if dropped > 0 {
		status.Dropped = dropped
	}

	application.Get().Event.Emit("recorder:status", status)
}

func parseFloat(s string, re *regexp.Regexp) float64 {
	m := re.FindStringSubmatch(s)
	if len(m) > 1 {
		var f float64
		fmt.Sscanf(m[1], "%f", &f)
		return f
	}
	return 0
}

func parseInt(s string, re *regexp.Regexp) int {
	m := re.FindStringSubmatch(s)
	if len(m) > 1 {
		var n int
		fmt.Sscanf(m[1], "%d", &n)
		return n
	}
	return 0
}
