# 摄像头录制功能实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 Extral 新增独立的摄像头视频录制功能，通过 FFmpeg 采集摄像头画面并编码保存，支持实时预览、音频录制和参数配置。

**Architecture:** 新增 `internal/recorder` 包作为独立 Wails Service，与现有抽帧模块解耦。Go 后端负责 FFmpeg 进程管理（预览帧流、录制控制、状态解析），前端通过 Wails 事件接收预览帧和录制状态。工具栏 Tab 切换两个功能模块。

**Tech Stack:** Go (Wails v3 Service) + React (hooks + components) + FFmpeg (avfoundation/dshow 采集 + libx264 编码)

---

## File Structure

### New Files

| File | Responsibility |
|------|---------------|
| `internal/recorder/models.go` | Camera, AudioDevice, RecordingConfig, RecordingStatus 数据结构 |
| `internal/recorder/camera.go` | FFprobe/原生设备枚举，平台特定解析 |
| `internal/recorder/camera_test.go` | 设备枚举解析测试 |
| `internal/recorder/builder.go` | FFmpeg 预览/录制命令参数构建 |
| `internal/recorder/builder_test.go` | 命令构建测试 |
| `internal/recorder/preview.go` | 预览帧流：启动 FFmpeg、读取 JPEG 帧、推送事件 |
| `internal/recorder/session.go` | 录制会话：开始/暂停/停止、状态解析 |
| `internal/recorder/service.go` | Wails Service 入口，整合所有组件 |
| `frontend/src/hooks/useRecorder.ts` | React hook：录制状态管理 + Wails 事件监听 |
| `frontend/src/components/RecorderPanel.tsx` | 录制 UI 面板：左右分栏布局 |

### Modified Files

| File | Change |
|------|--------|
| `main.go` | 注册 recorder service |
| `frontend/src/App.tsx` | Tab 切换、集成 RecorderPanel |
| `frontend/src/App.css` | 录制相关样式 |
| `.github/workflows/build-ffmpeg.yml` | 新增 avfoundation/dshow/libx264/aac 等编译选项 |

---

## Task 1: 数据模型

**Files:**
- Create: `internal/recorder/models.go`

- [ ] **Step 1: 创建 models.go**

```go
package recorder

import "time"

// Camera represents a video capture device.
type Camera struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// AudioDevice represents an audio capture device.
type AudioDevice struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// RecordingConfig holds all parameters for a recording session.
type RecordingConfig struct {
	CameraID    string `json:"cameraId"`
	AudioDevice string `json:"audioDevice"`
	Resolution  string `json:"resolution"`
	FPS         int    `json:"fps"`
	Codec       string `json:"codec"`
	RateControl string `json:"rateControl"` // "crf" | "bitrate"
	Quality     string `json:"quality"`     // "low" | "medium" | "high" | "custom"
	CRF         int    `json:"crf"`
	Bitrate     int    `json:"bitrate"`
	Format      string `json:"format"`
	MaxDuration int    `json:"maxDuration"`
	OutputDir   string `json:"outputDir"`
	FileName    string `json:"fileName"`
}

// RecordingStatus represents the current state of a recording session.
type RecordingStatus struct {
	State    string  `json:"state"`    // "idle" | "previewing" | "recording" | "paused"
	Duration int64   `json:"duration"` // seconds
	FileSize int64   `json:"fileSize"` // bytes
	FPS      float64 `json:"fps"`
	Dropped  int     `json:"dropped"`
	Error    string  `json:"error"`
}

// QualityToCRF maps quality presets to CRF values for libx264.
func QualityToCRF(quality string) int {
	switch quality {
	case "low":
		return 28
	case "medium":
		return 23
	case "high":
		return 18
	default:
		return 23
	}
}

// DefaultRecordingConfig returns a sensible default configuration.
func DefaultRecordingConfig() RecordingConfig {
	return RecordingConfig{
		Resolution:  "1280x720",
		FPS:         30,
		Codec:       "libx264",
		RateControl: "crf",
		Quality:     "medium",
		CRF:         23,
		Format:      "mp4",
	}
}

// GenerateFileName creates a timestamp-based filename.
func GenerateFileName(format string) string {
	ext := "mp4"
	if format == "mkv" {
		ext = "mkv"
	}
	return "recording_" + time.Now().Format("2006-01-02_15-04-05") + "." + ext
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/recorder/models.go
git commit -m "feat(recorder): add data models for camera recording"
```

---

## Task 2: FFmpeg 命令构建器 + 测试

**Files:**
- Create: `internal/recorder/builder.go`
- Create: `internal/recorder/builder_test.go`

- [ ] **Step 1: 编写 builder 测试**

```go
package recorder

import (
	"runtime"
	"testing"
)

func TestBuildPreviewArgs(t *testing.T) {
	camera := Camera{ID: "0", Name: "FaceTime HD Camera"}
	args := BuildPreviewArgs(camera)

	if runtime.GOOS == "darwin" {
		found := false
		for i, a := range args {
			if a == "-f" && i+1 < len(args) && args[i+1] == "avfoundation" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected -f avfoundation for darwin")
		}
	}
}

func TestBuildRecordingArgs_CRF(t *testing.T) {
	cfg := RecordingConfig{
		CameraID:    "0",
		AudioDevice: "",
		Resolution:  "1280x720",
		FPS:         30,
		Codec:       "libx264",
		RateControl: "crf",
		Quality:     "medium",
		CRF:         23,
		Format:      "mp4",
		MaxDuration: 0,
	}
	camera := Camera{ID: "0", Name: "FaceTime HD Camera"}
	args := BuildRecordingArgs(cfg, camera, "/tmp/out.mp4")

	joined := joinArgs(args)
	if !contains(joined, "-c:v") || !contains(joined, "libx264") {
		t.Errorf("expected -c:v libx264 in args, got %s", joined)
	}
	if !contains(joined, "-crf") || !contains(joined, "23") {
		t.Errorf("expected -crf 23 in args, got %s", joined)
	}
	if contains(joined, "-b:v") {
		t.Error("did not expect -b:v in CRF mode")
	}
}

func TestBuildRecordingArgs_Bitrate(t *testing.T) {
	cfg := RecordingConfig{
		CameraID:    "0",
		AudioDevice: "1",
		Resolution:  "1920x1080",
		FPS:         60,
		Codec:       "libx264",
		RateControl: "bitrate",
		Bitrate:     8000,
		Format:      "mp4",
	}
	camera := Camera{ID: "0", Name: "FaceTime HD Camera"}
	args := BuildRecordingArgs(cfg, camera, "/tmp/out.mp4")

	joined := joinArgs(args)
	if !contains(joined, "-b:v") || !contains(joined, "8000k") {
		t.Errorf("expected -b:v 8000k in args, got %s", joined)
	}
	if contains(joined, "-crf") {
		t.Error("did not expect -crf in bitrate mode")
	}
}

func TestBuildRecordingArgs_MaxDuration(t *testing.T) {
	cfg := RecordingConfig{
		CameraID:    "0",
		Resolution:  "1280x720",
		FPS:         30,
		Codec:       "libx264",
		RateControl: "crf",
		CRF:         23,
		Format:      "mp4",
		MaxDuration: 120,
	}
	camera := Camera{ID: "0", Name: "FaceTime HD Camera"}
	args := BuildRecordingArgs(cfg, camera, "/tmp/out.mp4")

	joined := joinArgs(args)
	if !contains(joined, "-t") || !contains(joined, "120") {
		t.Errorf("expected -t 120 in args, got %s", joined)
	}
}

func TestQualityToCRF(t *testing.T) {
	tests := []struct {
		quality string
		want    int
	}{
		{"low", 28},
		{"medium", 23},
		{"high", 18},
		{"unknown", 23},
	}
	for _, tt := range tests {
		got := QualityToCRF(tt.quality)
		if got != tt.want {
			t.Errorf("QualityToCRF(%q) = %d, want %d", tt.quality, got, tt.want)
		}
	}
}

func joinArgs(args []string) string {
	result := ""
	for _, a := range args {
		result += a + " "
	}
	return result
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsSub(s, sub))
}

func containsSub(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd /Users/wangjin/Downloads/video-extractor && go test ./internal/recorder/... -run TestBuild -v`
Expected: 编译失败，`BuildPreviewArgs` 和 `BuildRecordingArgs` 未定义

- [ ] **Step 3: 实现 builder.go**

```go
package recorder

import (
	"fmt"
	"runtime"
	"strings"
)

// BuildPreviewArgs constructs FFmpeg arguments for low-fps JPEG preview streaming.
func BuildPreviewArgs(camera Camera) []string {
	var args []string

	if runtime.GOOS == "darwin" {
		args = append(args,
			"-f", "avfoundation",
			"-framerate", "5",
			"-i", camera.ID,
		)
	} else {
		// Windows: dshow uses device name
		args = append(args,
			"-f", "dshow",
			"-i", "video="+camera.Name,
		)
	}

	args = append(args,
		"-vf", "scale=640:-1",
		"-f", "image2pipe",
		"-vcodec", "mjpeg",
		"pipe:1",
	)
	return args
}

// BuildRecordingArgs constructs FFmpeg arguments for video recording.
func BuildRecordingArgs(cfg RecordingConfig, camera Camera, outputPath string) []string {
	var args []string

	// Input device
	if runtime.GOOS == "darwin" {
		input := cfg.CameraID
		if cfg.AudioDevice != "" && cfg.AudioDevice != "none" {
			input = cfg.CameraID + ":" + cfg.AudioDevice
		} else {
			input = cfg.CameraID + ":none"
		}
		args = append(args,
			"-f", "avfoundation",
			"-framerate", fmt.Sprintf("%d", cfg.FPS),
			"-i", input,
		)
	} else {
		// Windows: dshow
		input := "video=" + camera.Name
		if cfg.AudioDevice != "" && cfg.AudioDevice != "none" {
			input += ":audio=" + cfg.AudioDevice
		}
		args = append(args,
			"-f", "dshow",
			"-framerate", fmt.Sprintf("%d", cfg.FPS),
			"-i", input,
		)
	}

	// Resolution scaling
	if cfg.Resolution != "" {
		parts := strings.Split(cfg.Resolution, "x")
		if len(parts) == 2 {
			args = append(args, "-vf", fmt.Sprintf("scale=%s:%s", parts[0], parts[1]))
		}
	}

	// Video codec
	args = append(args, "-c:v", cfg.Codec)

	// Rate control
	if cfg.RateControl == "bitrate" && cfg.Bitrate > 0 {
		args = append(args, "-b:v", fmt.Sprintf("%dk", cfg.Bitrate))
	} else {
		crf := cfg.CRF
		if cfg.Quality != "custom" {
			crf = QualityToCRF(cfg.Quality)
		}
		args = append(args, "-crf", fmt.Sprintf("%d", crf))
	}

	args = append(args, "-preset", "medium")

	// Audio codec (only if audio device selected)
	if cfg.AudioDevice != "" && cfg.AudioDevice != "none" {
		args = append(args, "-c:a", "aac", "-b:a", "128k")
	}

	// Max duration
	if cfg.MaxDuration > 0 {
		args = append(args, "-t", fmt.Sprintf("%d", cfg.MaxDuration))
	}

	args = append(args, "-y", outputPath)
	return args
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `cd /Users/wangjin/Downloads/video-extractor && go test ./internal/recorder/... -v`
Expected: 所有测试 PASS

- [ ] **Step 5: Commit**

```bash
git add internal/recorder/builder.go internal/recorder/builder_test.go
git commit -m "feat(recorder): add FFmpeg command builder with tests"
```

---

## Task 3: 设备枚举 + 测试

**Files:**
- Create: `internal/recorder/camera.go`
- Create: `internal/recorder/camera_test.go`

- [ ] **Step 1: 编写设备枚举测试**

```go
package recorder

import (
	"testing"
)

func TestParseAVFoundationDevices_VideoAndAudio(t *testing.T) {
	output := `[AVFoundation indev @ 0x1] AVFoundation video devices:
[AVFoundation indev @ 0x1] [0] FaceTime HD Camera (Built-in)
[AVFoundation indev @ 0x1] [1] OBS Virtual Camera
[AVFoundation indev @ 0x1] AVFoundation audio devices:
[AVFoundation indev @ 0x1] [0] MacBook Pro Microphone (Built-in)
[AVFoundation indev @ 0x1] [1] External Microphone
`

	cameras, audios := parseAVFoundationDevices(output)
	if len(cameras) != 2 {
		t.Fatalf("expected 2 cameras, got %d", len(cameras))
	}
	if cameras[0].ID != "0" || cameras[0].Name != "FaceTime HD Camera (Built-in)" {
		t.Errorf("camera[0] unexpected: %+v", cameras[0])
	}
	if cameras[1].ID != "1" || cameras[1].Name != "OBS Virtual Camera" {
		t.Errorf("camera[1] unexpected: %+v", cameras[1])
	}
	if len(audios) != 2 {
		t.Fatalf("expected 2 audio devices, got %d", len(audios))
	}
	if audios[0].ID != "0" || audios[0].Name != "MacBook Pro Microphone (Built-in)" {
		t.Errorf("audio[0] unexpected: %+v", audios[0])
	}
}

func TestParseAVFoundationDevices_Empty(t *testing.T) {
	output := `[AVFoundation indev @ 0x1] AVFoundation video devices:
[AVFoundation indev @ 0x1] AVFoundation audio devices:
`
	cameras, audios := parseAVFoundationDevices(output)
	if len(cameras) != 0 {
		t.Errorf("expected 0 cameras, got %d", len(cameras))
	}
	if len(audios) != 0 {
		t.Errorf("expected 0 audio devices, got %d", len(audios))
	}
}

func TestParseDShowDevices(t *testing.T) {
	output := `[dshow @ 0x1] "DirectShow video devices"
[dshow @ 0x1]  Alternative name "@device_pnp_..."
[dshow @ 0x1]  [Integrated Webcam]
[dshow @ 0x1]  Alternative name "@device_pnp_..."
[dshow @ 0x1] "DirectShow audio devices"
[dshow @ 0x1]  [Microphone Array (Realtek Audio)]
[dshow @ 0x1]  [External Mic]
`

	cameras, audios := parseDShowDevices(output)
	if len(cameras) != 1 {
		t.Fatalf("expected 1 camera, got %d", len(cameras))
	}
	if cameras[0].Name != "Integrated Webcam" {
		t.Errorf("camera[0] unexpected: %+v", cameras[0])
	}
	if len(audios) != 2 {
		t.Fatalf("expected 2 audio devices, got %d", len(audios))
	}
	if audios[0].Name != "Microphone Array (Realtek Audio)" {
		t.Errorf("audio[0] unexpected: %+v", audios[0])
	}
}

func TestParseSystemProfiler(t *testing.T) {
	output := `Camera:
    Camera Model:

        FaceTime HD Camera (Built-in):

          Unique ID: 0x12345678

    Camera Model:

        OBS Virtual Camera:

          Unique ID: 0x87654321
`

	cameras := parseSystemProfiler(output)
	if len(cameras) != 2 {
		t.Fatalf("expected 2 cameras, got %d", len(cameras))
	}
	if cameras[0].Name != "FaceTime HD Camera (Built-in)" {
		t.Errorf("camera[0] unexpected: %+v", cameras[0])
	}
	if cameras[1].Name != "OBS Virtual Camera" {
		t.Errorf("camera[1] unexpected: %+v", cameras[1])
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd /Users/wangjin/Downloads/video-extractor && go test ./internal/recorder/... -run TestParse -v`
Expected: 编译失败，解析函数未定义

- [ ] **Step 3: 实现 camera.go**

```go
package recorder

import (
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"video-extractor/internal/ffmpeg"
)

var (
	avfDeviceRe  = regexp.MustCompile(`\[(\d+)\]\s+(.+)`)
	dshowDeviceRe = regexp.MustCompile(`\[(.+?)\]`)
	profilerNameRe = regexp.MustCompile(`^\s{4,8}(\S.+):$`)
)

// ListCameras enumerates available video capture devices.
func ListCameras(ffmpegMgr *ffmpeg.Manager) ([]Camera, error) {
	if runtime.GOOS == "darwin" {
		return listCamerasDarwin(ffmpegMgr)
	}
	return listCamerasWindows(ffmpegMgr)
}

// ListAudioDevices enumerates available audio capture devices.
func ListAudioDevices(ffmpegMgr *ffmpeg.Manager) ([]AudioDevice, error) {
	if runtime.GOOS == "darwin" {
		return listAudioDarwin(ffmpegMgr)
	}
	return listAudioWindows(ffmpegMgr)
}

func listCamerasDarwin(ffmpegMgr *ffmpeg.Manager) ([]Camera, error) {
	out, err := ffmpegMgr.RunFFprobe("-list_devices", "true", "-f", "avfoundation", "-i", "dummy")
	if err == nil {
		cameras, _ := parseAVFoundationDevices(string(out))
		if len(cameras) > 0 {
			return cameras, nil
		}
	}
	// Fallback: system_profiler
	return listCamerasSystemProfiler(), nil
}

func listAudioDarwin(ffmpegMgr *ffmpeg.Manager) ([]AudioDevice, error) {
	out, err := ffmpegMgr.RunFFprobe("-list_devices", "true", "-f", "avfoundation", "-i", "dummy")
	if err == nil {
		_, audios := parseAVFoundationDevices(string(out))
		return audios, nil
	}
	return nil, nil
}

func listCamerasWindows(ffmpegMgr *ffmpeg.Manager) ([]Camera, error) {
	out, err := ffmpegMgr.RunFFprobe("-list_devices", "true", "-f", "dshow", "-i", "dummy")
	if err == nil {
		cameras, _ := parseDShowDevices(string(out))
		return cameras, nil
	}
	return nil, nil
}

func listAudioWindows(ffmpegMgr *ffmpeg.Manager) ([]AudioDevice, error) {
	out, err := ffmpegMgr.RunFFprobe("-list_devices", "true", "-f", "dshow", "-i", "dummy")
	if err == nil {
		_, audios := parseDShowDevices(string(out))
		return audios, nil
	}
	return nil, nil
}

func listCamerasSystemProfiler() []Camera {
	out, err := exec.Command("system_profiler", "SPCameraDataType").Output()
	if err != nil {
		return nil
	}
	return parseSystemProfiler(string(out))
}

// parseAVFoundationDevices parses FFprobe avfoundation device listing output.
func parseAVFoundationDevices(output string) (cameras []Camera, audios []AudioDevice) {
	section := ""
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "AVFoundation video devices:") {
			section = "video"
			continue
		}
		if strings.Contains(trimmed, "AVFoundation audio devices:") {
			section = "audio"
			continue
		}
		m := avfDeviceRe.FindStringSubmatch(trimmed)
		if m == nil {
			continue
		}
		id := m[1]
		name := strings.TrimSpace(m[2])
		if section == "video" {
			cameras = append(cameras, Camera{ID: id, Name: name})
		} else if section == "audio" {
			audios = append(audios, AudioDevice{ID: id, Name: name})
		}
	}
	return
}

// parseDShowDevices parses FFprobe dshow device listing output.
func parseDShowDevices(output string) (cameras []Camera, audios []AudioDevice) {
	section := ""
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "DirectShow video devices") {
			section = "video"
			continue
		}
		if strings.Contains(trimmed, "DirectShow audio devices") {
			section = "audio"
			continue
		}
		// Skip "Alternative name" lines
		if strings.Contains(trimmed, "Alternative name") {
			continue
		}
		m := dshowDeviceRe.FindStringSubmatch(trimmed)
		if m == nil {
			continue
		}
		name := m[1]
		if section == "video" {
			cameras = append(cameras, Camera{ID: name, Name: name})
		} else if section == "audio" {
			audios = append(audios, AudioDevice{ID: name, Name: name})
		}
	}
	return
}

// parseSystemProfiler parses macOS system_profiler SPCameraDataType output.
func parseSystemProfiler(output string) []Camera {
	var cameras []Camera
	idx := 0
	for _, line := range strings.Split(output, "\n") {
		m := profilerNameRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		name := strings.TrimSuffix(strings.TrimSpace(m[1]), ":")
		if name == "" {
			continue
		}
		cameras = append(cameras, Camera{ID: fmt.Sprintf("%d", idx), Name: name})
		idx++
	}
	return cameras
}
```

**注意**：需要在文件顶部 import 中加入 `"fmt"`。

- [ ] **Step 4: 运行测试确认通过**

Run: `cd /Users/wangjin/Downloads/video-extractor && go test ./internal/recorder/... -run TestParse -v`
Expected: 所有测试 PASS

- [ ] **Step 5: 运行全部测试**

Run: `cd /Users/wangjin/Downloads/video-extractor && go test ./internal/recorder/... -v`
Expected: 所有测试 PASS

- [ ] **Step 6: Commit**

```bash
git add internal/recorder/camera.go internal/recorder/camera_test.go
git commit -m "feat(recorder): add camera/audio device enumeration with platform support"
```

---

## Task 4: 预览帧流

**Files:**
- Create: `internal/recorder/preview.go`

- [ ] **Step 1: 实现 preview.go**

```go
package recorder

import (
	"context"
	"encoding/base64"
	"io"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"

	"video-extractor/internal/ffmpeg"
)

// Preview manages a low-fps MJPEG preview stream from a camera.
type Preview struct {
	mu       sync.Mutex
	ffmpeg   *ffmpeg.Manager
	cmd      *exec.Cmd
	cancel   context.CancelFunc
	camera   Camera
}

// NewPreview creates a new Preview instance.
func NewPreview(ffmpegMgr *ffmpeg.Manager) *Preview {
	return &Preview{ffmpeg: ffmpegMgr}
}

// Start begins streaming preview frames from the given camera.
// Frames are emitted as base64 JPEG via the "recorder:preview-frame" event.
func (p *Preview) Start(camera Camera) error {
	p.Stop()

	ctx, cancel := context.WithCancel(context.Background())

	args := BuildPreviewArgs(camera)
	cmd := p.ffmpeg.RunFFmpeg(args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return err
	}
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		cancel()
		return err
	}

	p.mu.Lock()
	p.cmd = cmd
	p.cancel = cancel
	p.camera = camera
	p.mu.Unlock()

	// Read stderr for error detection
	go func() {
		buf := make([]byte, 4096)
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			n, err := stderr.Read(buf)
			if err != nil {
				return
			}
			line := string(buf[:n])
			application.Get().Event.Emit("recorder:log", line)
		}
	}()

	// Read JPEG frames from stdout
	go p.readFrames(ctx, stdout)

	return nil
}

// Stop terminates the preview stream.
func (p *Preview) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cancel != nil {
		p.cancel()
		p.cancel = nil
	}
	if p.cmd != nil && p.cmd.Process != nil {
		p.cmd.Process.Kill()
		p.cmd.Wait()
	}
	p.cmd = nil
}

// IsRunning returns whether preview is currently active.
func (p *Preview) IsRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.cmd != nil && p.cmd.Process != nil
}

// readFrames reads JPEG frames from FFmpeg's stdout and emits them as events.
func (p *Preview) readFrames(ctx context.Context, stdout io.Reader) {
	buf := make([]byte, 64*1024)
	var frameBuf []byte
	inFrame := false

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		n, err := stdout.Read(buf)
		if n > 0 {
			frameBuf = append(frameBuf, buf[:n]...)
		}
		if err != nil {
			if err != io.EOF {
				application.Get().Event.Emit("recorder:error",
					map[string]string{"error": err.Error()})
			}
			return
		}

		// Extract complete JPEG frames
		for {
			if !inFrame {
				idx := findJPEGStart(frameBuf)
				if idx == -1 {
					// Keep last few bytes in case marker spans reads
					if len(frameBuf) > 4 {
						frameBuf = frameBuf[len(frameBuf)-4:]
					}
					break
				}
				frameBuf = frameBuf[idx:]
				inFrame = true
			}

			if inFrame {
				endIdx := findJPEGEnd(frameBuf)
				if endIdx == -1 {
					break // Need more data
				}
				frame := make([]byte, endIdx+2)
				copy(frame, frameBuf[:endIdx+2])
				frameBuf = frameBuf[endIdx+2:]
				inFrame = false

				// Emit base64 encoded frame
				encoded := base64.StdEncoding.EncodeToString(frame)
				application.Get().Event.Emit("recorder:preview-frame", encoded)
			}
		}
	}
}

// findJPEGStart finds the index of JPEG start marker (0xFF 0xD8).
func findJPEGStart(data []byte) int {
	for i := 0; i < len(data)-1; i++ {
		if data[i] == 0xFF && data[i+1] == 0xD8 {
			return i
		}
	}
	return -1
}

// findJPEGEnd finds the index of JPEG end marker (0xFF 0xD9).
func findJPEGEnd(data []byte) int {
	for i := 0; i < len(data)-1; i++ {
		if data[i] == 0xFF && data[i+1] == 0xD9 {
			return i
		}
	}
	return -1
}
```

**注意**：需要在 import 中加入 `"os/exec"`。

- [ ] **Step 2: Commit**

```bash
git add internal/recorder/preview.go
git commit -m "feat(recorder): add preview frame streaming with JPEG pipe reader"
```

---

## Task 5: 录制会话

**Files:**
- Create: `internal/recorder/session.go`

- [ ] **Step 1: 实现 session.go**

```go
package recorder

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"

	"video-extractor/internal/ffmpeg"
)

// Session manages a recording session lifecycle.
type Session struct {
	mu       sync.Mutex
	ffmpeg   *ffmpeg.Manager
	cmd      *exec.Cmd
	cancel   context.CancelFunc
	config   RecordingConfig
	camera   Camera
	started  time.Time
	paused   bool
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
	// Need to import "io" at the top
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
		state = "idle"
		errMsg = fmt.Sprintf("FFmpeg 退出异常: %v", waitErr)
		application.Get().Event.Emit("recorder:error",
			map[string]string{"error": errMsg})
	}

	application.Get().Event.Emit("recorder:status", RecordingStatus{
		State:    state,
		Duration: int64(time.Since(s.started).Seconds()),
		Error:    errMsg,
	})
	if state == "idle" && errMsg == "" {
		application.Get().Event.Emit("recorder:log", "[完成]")
	}
}

// emitStatusFromLine parses an FFmpeg progress line and emits a status event.
func (s *Session) emitStatusFromLine(line string) {
	if !strings.Contains(line, "frame=") {
		return
	}

	frame, _, _ := ffmpeg.ParseProgress(line)
	fps := parseFloat(line, `fps=\s*([\d.]+)`)
	sizeKB := parseFloat(line, `size=\s*(\d+)kB`)

	status := RecordingStatus{
		State:    "recording",
		Duration: int64(time.Since(s.started).Seconds()),
		FPS:      fps,
	}
	if s.IsPaused() {
		status.State = "paused"
	}
	if frame > 0 {
		status.FPS = fps
	}
	if sizeKB > 0 {
		status.FileSize = int64(sizeKB * 1024)
	}

	application.Get().Event.Emit("recorder:status", status)
}

// parseFloat extracts a float value from a string using a regex pattern.
func parseFloat(s string, pattern string) float64 {
	re := regexp.MustCompile(pattern)
	m := re.FindStringSubmatch(s)
	if len(m) > 1 {
		var f float64
		fmt.Sscanf(m[1], "%f", &f)
		return f
	}
	return 0
}
```

**注意**：需要在 import 中加入 `"io"`、`"regexp"`，并确保所有必要的包已导入。`findJPEGStart` 和 `findJPEGEnd` 不需要在此文件中，它们在 `preview.go` 中。

- [ ] **Step 2: Commit**

```bash
git add internal/recorder/session.go
git commit -m "feat(recorder): add recording session with start/stop/pause/resume"
```

---

## Task 6: Wails Service 层

**Files:**
- Create: `internal/recorder/service.go`

- [ ] **Step 1: 实现 service.go**

```go
package recorder

import (
	"context"
	"fmt"

	"github.com/wailsapp/wails/v3/pkg/application"

	"video-extractor/internal/ffmpeg"
)

// Service is the Wails service entry point for camera recording.
type Service struct {
	ctx       context.Context
	ffmpegMgr *ffmpeg.Manager
	preview   *Preview
	session   *Session
}

// NewService creates a new recorder Service.
func NewService() *Service {
	mgr := ffmpeg.NewManager()
	return &Service{
		ffmpegMgr: mgr,
		preview:   NewPreview(mgr),
		session:   NewSession(mgr),
	}
}

// ServiceStartup is called when the Wails application starts.
func (s *Service) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	s.ctx = ctx
	return nil
}

// ServiceShutdown is called when the Wails application shuts down.
func (s *Service) ServiceShutdown() error {
	s.preview.Stop()
	s.session.Stop()
	return nil
}

// ListCameras returns available video capture devices.
func (s *Service) ListCameras() ([]Camera, error) {
	return ListCameras(s.ffmpegMgr)
}

// ListAudioDevices returns available audio capture devices.
func (s *Service) ListAudioDevices() ([]AudioDevice, error) {
	return ListAudioDevices(s.ffmpegMgr)
}

// StartPreview begins streaming preview frames from the specified camera.
func (s *Service) StartPreview(cameraID string) error {
	cameras, err := s.ListCameras()
	if err != nil {
		return fmt.Errorf("获取摄像头列表失败: %w", err)
	}
	var camera Camera
	for _, c := range cameras {
		if c.ID == cameraID {
			camera = c
			break
		}
	}
	if camera.ID == "" {
		return fmt.Errorf("未找到摄像头: %s", cameraID)
	}
	return s.preview.Start(camera)
}

// StopPreview stops the current preview stream.
func (s *Service) StopPreview() error {
	s.preview.Stop()
	return nil
}

// StartRecording begins a recording session with the given configuration.
func (s *Service) StartRecording(config RecordingConfig) error {
	if s.session.IsRunning() {
		return fmt.Errorf("已有录制正在进行")
	}

	// Stop preview to release the camera device
	s.preview.Stop()

	cameras, err := s.ListCameras()
	if err != nil {
		return fmt.Errorf("获取摄像头列表失败: %w", err)
	}
	var camera Camera
	for _, c := range cameras {
		if c.ID == config.CameraID {
			camera = c
			break
		}
	}
	if camera.ID == "" {
		return fmt.Errorf("未找到摄像头: %s", config.CameraID)
	}

	// Auto-generate filename if not set
	if config.FileName == "" {
		config.FileName = GenerateFileName(config.Format)
	}

	return s.session.Start(config, camera)
}

// StopRecording stops the current recording session.
func (s *Service) StopRecording() error {
	s.session.Stop()
	return nil
}

// PauseRecording pauses the current recording (macOS/Linux only).
func (s *Service) PauseRecording() error {
	return s.session.Pause()
}

// ResumeRecording resumes a paused recording (macOS/Linux only).
func (s *Service) ResumeRecording() error {
	return s.session.Resume()
}

// SelectOutputDir opens a directory picker dialog.
func (s *Service) SelectOutputDir() string {
	result, err := application.Get().Dialog.OpenFile().
		SetTitle("选择录制保存路径").
		CanChooseDirectories(true).
		CanChooseFiles(false).
		PromptForSingleSelection()
	if err != nil || result == "" {
		return ""
	}
	return result
}

// IsPreviewing returns whether a preview is currently active.
func (s *Service) IsPreviewing() bool {
	return s.preview.IsRunning()
}

// IsRecording returns whether a recording is currently active.
func (s *Service) IsRecording() bool {
	return s.session.IsRunning()
}
```

- [ ] **Step 2: 运行全部 recorder 测试**

Run: `cd /Users/wangjin/Downloads/video-extractor && go test ./internal/recorder/... -v`
Expected: 所有测试 PASS

- [ ] **Step 3: Commit**

```bash
git add internal/recorder/service.go
git commit -m "feat(recorder): add Wails service layer with full API"
```

---

## Task 7: main.go 注册 + 生成绑定

**Files:**
- Modify: `main.go`

- [ ] **Step 1: 在 main.go 中注册 recorder service**

在 import 块中添加：
```go
"video-extractor/internal/recorder"
```

在 `func main()` 中，`appService := app.NewService(version)` 之后添加：
```go
recorderService := recorder.NewService()
```

在 `Services` 列表中添加：
```go
application.NewService(recorderService),
```

- [ ] **Step 2: 确保项目编译通过**

Run: `cd /Users/wangjin/Downloads/video-extractor && go build ./...`
Expected: 编译成功

- [ ] **Step 3: 生成 Wails 前端绑定**

Run: `cd /Users/wangjin/Downloads/video-extractor && wails3 generate bindings`
Expected: 在 `frontend/bindings/video-extractor/internal/recorder/` 下生成 `service.ts` 和 `models.ts`

- [ ] **Step 4: 验证生成的绑定文件存在**

Run: `ls -la frontend/bindings/video-extractor/internal/recorder/`
Expected: 看到 `service.ts` 和 `models.ts` 文件

- [ ] **Step 5: Commit**

```bash
git add main.go frontend/bindings/
git commit -m "feat(recorder): register recorder service and generate bindings"
```

---

## Task 8: 前端 useRecorder Hook

**Files:**
- Create: `frontend/src/hooks/useRecorder.ts`

- [ ] **Step 1: 实现 useRecorder.ts**

参考 `frontend/bindings/video-extractor/internal/recorder/service.ts` 中生成的绑定方法名（在 Task 7 生成后查看确认），以下使用预期的方法名：

```typescript
import { useState, useEffect, useCallback } from 'react'
import { Events } from '@wailsio/runtime'
import * as RecorderService from '../../bindings/video-extractor/internal/recorder/service'
import { RecordingConfig } from '../../bindings/video-extractor/internal/recorder/models'

export interface Camera {
  id: string
  name: string
}

export interface AudioDevice {
  id: string
  name: string
}

export interface RecorderState {
  cameras: Camera[]
  audioDevices: AudioDevice[]
  selectedCamera: string
  selectedAudio: string
  config: RecordingConfig
  previewFrame: string
  isPreviewing: boolean
  isRecording: boolean
  isPaused: boolean
  duration: number
  fileSize: number
  fps: number
  dropped: number
  logs: string[]
  error: string | null
}

const defaultConfig: RecordingConfig = {
  cameraId: '',
  audioDevice: '',
  resolution: '1280x720',
  fps: 30,
  codec: 'libx264',
  rateControl: 'crf',
  quality: 'medium',
  crf: 23,
  bitrate: 5000,
  format: 'mp4',
  maxDuration: 0,
  outputDir: '',
  fileName: '',
}

export function useRecorder() {
  const [cameras, setCameras] = useState<Camera[]>([])
  const [audioDevices, setAudioDevices] = useState<AudioDevice[]>([])
  const [selectedCamera, setSelectedCamera] = useState('')
  const [selectedAudio, setSelectedAudio] = useState('')
  const [config, setConfig] = useState<RecordingConfig>(defaultConfig)
  const [previewFrame, setPreviewFrame] = useState('')
  const [isPreviewing, setIsPreviewing] = useState(false)
  const [isRecording, setIsRecording] = useState(false)
  const [isPaused, setIsPaused] = useState(false)
  const [duration, setDuration] = useState(0)
  const [fileSize, setFileSize] = useState(0)
  const [fps, setFps] = useState(0)
  const [dropped, setDropped] = useState(0)
  const [logs, setLogs] = useState<string[]>([])
  const [error, setError] = useState<string | null>(null)

  // Load devices and listen to events on mount
  useEffect(() => {
    loadDevices()

    const off1 = Events.On('recorder:preview-frame', (ev: any) => {
      setPreviewFrame(ev.data || ev)
    })

    const off2 = Events.On('recorder:status', (ev: any) => {
      const status = ev.data || ev
      if (status.state === 'recording') {
        setIsRecording(true)
        setIsPaused(false)
      } else if (status.state === 'paused') {
        setIsPaused(true)
      } else {
        setIsRecording(false)
        setIsPaused(false)
      }
      if (status.duration !== undefined) setDuration(status.duration)
      if (status.fileSize !== undefined) setFileSize(status.fileSize)
      if (status.fps !== undefined) setFps(status.fps)
      if (status.dropped !== undefined) setDropped(status.dropped)
    })

    const off3 = Events.On('recorder:log', (ev: any) => {
      const line = ev.data || ev
      setLogs(prev => [...prev, line])
    })

    const off4 = Events.On('recorder:error', (ev: any) => {
      const data = ev.data || ev
      setError(data.error || String(data))
    })

    return () => {
      off1()
      off2()
      off3()
      off4()
    }
  }, [])

  const loadDevices = useCallback(async () => {
    try {
      const cams = await RecorderService.ListCameras()
      setCameras(cams || [])
      if (cams && cams.length > 0 && !selectedCamera) {
        setSelectedCamera(cams[0].id)
      }
    } catch (e) {
      console.error('ListCameras failed:', e)
    }
    try {
      const devs = await RecorderService.ListAudioDevices()
      setAudioDevices(devs || [])
    } catch (e) {
      console.error('ListAudioDevices failed:', e)
    }
  }, [])

  // Start preview when camera is selected
  useEffect(() => {
    if (selectedCamera && !isRecording) {
      RecorderService.StartPreview(selectedCamera).then(() => {
        setIsPreviewing(true)
        setPreviewFrame('')
      }).catch(e => {
        console.error('StartPreview failed:', e)
        setError(String(e))
      })
    }
    return () => {
      // Stop preview when camera changes or component unmounts
    }
  }, [selectedCamera])

  const selectCamera = useCallback((id: string) => {
    setSelectedCamera(id)
    setConfig(prev => ({ ...prev, cameraId: id }))
  }, [])

  const selectAudio = useCallback((id: string) => {
    setSelectedAudio(id)
    setConfig(prev => ({ ...prev, audioDevice: id }))
  }, [])

  const updateConfig = useCallback((updates: Partial<RecordingConfig>) => {
    setConfig(prev => ({ ...prev, ...updates }))
  }, [])

  const startRecording = useCallback(async () => {
    setLogs([])
    setError(null)
    try {
      const fullConfig: RecordingConfig = {
        ...config,
        cameraId: selectedCamera,
        audioDevice: selectedAudio,
      }
      if (!fullConfig.outputDir) {
        const dir = await RecorderService.SelectOutputDir()
        if (dir) {
          fullConfig.outputDir = dir
          setConfig(prev => ({ ...prev, outputDir: dir }))
        }
      }
      await RecorderService.StartRecording(fullConfig)
      setIsRecording(true)
      setIsPreviewing(false)
    } catch (e) {
      setError(String(e))
    }
  }, [config, selectedCamera, selectedAudio])

  const stopRecording = useCallback(async () => {
    try {
      await RecorderService.StopRecording()
      setIsRecording(false)
      setIsPaused(false)
      // Restart preview
      if (selectedCamera) {
        RecorderService.StartPreview(selectedCamera).then(() => {
          setIsPreviewing(true)
        }).catch(() => {})
      }
    } catch (e) {
      setError(String(e))
    }
  }, [selectedCamera])

  const pauseRecording = useCallback(async () => {
    try {
      await RecorderService.PauseRecording()
      setIsPaused(true)
    } catch (e) {
      setError(String(e))
    }
  }, [])

  const resumeRecording = useCallback(async () => {
    try {
      await RecorderService.ResumeRecording()
      setIsPaused(false)
    } catch (e) {
      setError(String(e))
    }
  }, [])

  const selectOutputDir = useCallback(async () => {
    const dir = await RecorderService.SelectOutputDir()
    if (dir) {
      setConfig(prev => ({ ...prev, outputDir: dir }))
    }
  }, [])

  return {
    cameras,
    audioDevices,
    selectedCamera,
    selectedAudio,
    config,
    previewFrame,
    isPreviewing,
    isRecording,
    isPaused,
    duration,
    fileSize,
    fps,
    dropped,
    logs,
    error,
    selectCamera,
    selectAudio,
    updateConfig,
    startRecording,
    stopRecording,
    pauseRecording,
    resumeRecording,
    selectOutputDir,
    loadDevices,
  }
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/hooks/useRecorder.ts
git commit -m "feat(recorder): add useRecorder hook for frontend state management"
```

---

## Task 9: 前端 RecorderPanel 组件

**Files:**
- Create: `frontend/src/components/RecorderPanel.tsx`

- [ ] **Step 1: 实现 RecorderPanel.tsx**

```tsx
import { useRecorder } from '../hooks/useRecorder'

export default function RecorderPanel() {
  const {
    cameras, audioDevices,
    selectedCamera, selectedAudio,
    config, previewFrame,
    isPreviewing, isRecording, isPaused,
    duration, fileSize, fps,
    logs, error,
    selectCamera, selectAudio,
    updateConfig,
    startRecording, stopRecording,
    pauseRecording, resumeRecording,
    selectOutputDir,
  } = useRecorder()

  const formatDuration = (sec: number) => {
    const h = Math.floor(sec / 3600)
    const m = Math.floor((sec % 3600) / 60)
    const s = Math.floor(sec % 60)
    return `${h.toString().padStart(2, '0')}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`
  }

  const formatSize = (bytes: number) => {
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
    return `${(bytes / 1024 / 1024).toFixed(1)} MB`
  }

  return (
    <div className="recorder-panel">
      {/* Left: Preview + Controls */}
      <div className="recorder-left">
        <div className="recorder-preview">
          {isRecording ? (
            <div className="recorder-recording-overlay">
              <div className="recording-indicator">
                <span className="recording-dot" />
                {isPaused ? '已暂停' : '录制中'}
              </div>
              <div className="recording-time">{formatDuration(duration)}</div>
            </div>
          ) : previewFrame ? (
            <img src={`data:image/jpeg;base64,${previewFrame}`} alt="预览" />
          ) : (
            <div className="recorder-preview-empty">
              <span className="preview-icon">📷</span>
              <span>{cameras.length === 0 ? '未检测到摄像头设备' : '选择摄像头开始预览'}</span>
            </div>
          )}
        </div>

        <div className="recorder-controls">
          {!isRecording ? (
            <button
              className="btn btn-primary recorder-btn-record"
              disabled={!selectedCamera}
              onClick={startRecording}
            >
              ● 开始录制
            </button>
          ) : (
            <>
              {isPaused ? (
                <button className="btn btn-secondary" onClick={resumeRecording}>
                  ▶ 继续
                </button>
              ) : (
                <button className="btn btn-secondary" onClick={pauseRecording}>
                  ⏸ 暂停
                </button>
              )}
              <button className="btn btn-danger" onClick={stopRecording}>
                ⏹ 停止
              </button>
            </>
          )}
        </div>

        <div className="recorder-status">
          <span>⏱ {formatDuration(duration)}</span>
          <span>📁 {formatSize(fileSize)}</span>
          <span>🎬 {fps.toFixed(0)} fps</span>
        </div>
      </div>

      {/* Right: Config Panel */}
      <div className="recorder-right">
        <div className="config-section">
          <label className="config-label">摄像头</label>
          <select
            className="recorder-select"
            value={selectedCamera}
            onChange={e => selectCamera(e.target.value)}
            disabled={isRecording}
          >
            <option value="">选择摄像头</option>
            {cameras.map(c => (
              <option key={c.id} value={c.id}>{c.name}</option>
            ))}
          </select>
        </div>

        <div className="config-section">
          <label className="config-label">麦克风</label>
          <select
            className="recorder-select"
            value={selectedAudio}
            onChange={e => selectAudio(e.target.value)}
            disabled={isRecording}
          >
            <option value="">不录制音频</option>
            {audioDevices.map(d => (
              <option key={d.id} value={d.id}>{d.name}</option>
            ))}
          </select>
        </div>

        <div className="config-section">
          <label className="config-label">分辨率</label>
          <div className="btn-group">
            {[
              { label: '1080p', value: '1920x1080' },
              { label: '720p', value: '1280x720' },
              { label: '480p', value: '854x480' },
            ].map(r => (
              <button
                key={r.value}
                className={`btn-option ${config.resolution === r.value ? 'active' : ''}`}
                disabled={isRecording}
                onClick={() => updateConfig({ resolution: r.value })}
              >{r.label}</button>
            ))}
          </div>
        </div>

        <div className="config-section">
          <label className="config-label">帧率</label>
          <div className="btn-group">
            {[30, 60].map(f => (
              <button
                key={f}
                className={`btn-option ${config.fps === f ? 'active' : ''}`}
                disabled={isRecording}
                onClick={() => updateConfig({ fps: f })}
              >{f} fps</button>
            ))}
          </div>
        </div>

        <div className="config-section">
          <label className="config-label">质量控制</label>
          <div className="btn-group" style={{ marginBottom: '6px' }}>
            {['crf', 'bitrate'].map(mode => (
              <button
                key={mode}
                className={`btn-option ${config.rateControl === mode ? 'active' : ''}`}
                disabled={isRecording}
                onClick={() => updateConfig({ rateControl: mode })}
              >{mode === 'crf' ? 'CRF 质量' : '码率'}</button>
            ))}
          </div>
          {config.rateControl === 'crf' ? (
            <>
              <div className="btn-group" style={{ marginBottom: '6px' }}>
                {[
                  { label: '低', value: 'low' },
                  { label: '中', value: 'medium' },
                  { label: '高', value: 'high' },
                  { label: '自定义', value: 'custom' },
                ].map(q => (
                  <button
                    key={q.value}
                    className={`btn-option ${config.quality === q.value ? 'active' : ''}`}
                    disabled={isRecording}
                    onClick={() => updateConfig({ quality: q.value })}
                  >{q.label}</button>
                ))}
              </div>
              {config.quality === 'custom' && (
                <div className="param-row">
                  <span className="param-label">CRF</span>
                  <input
                    type="range" min="0" max="51"
                    value={config.crf}
                    onChange={e => updateConfig({ crf: Number(e.target.value) })}
                  />
                  <span className="param-value">{config.crf}</span>
                </div>
              )}
            </>
          ) : (
            <div className="param-row">
              <input
                type="number" className="input-sm"
                value={config.bitrate}
                onChange={e => updateConfig({ bitrate: Number(e.target.value) })}
                disabled={isRecording}
              />
              <span className="param-unit">kbps</span>
            </div>
          )}
        </div>

        <div className="config-section">
          <label className="config-label">输出格式</label>
          <div className="btn-group">
            {['mp4', 'mkv'].map(f => (
              <button
                key={f}
                className={`btn-option ${config.format === f ? 'active' : ''}`}
                disabled={isRecording}
                onClick={() => updateConfig({ format: f })}
              >{f.toUpperCase()}</button>
            ))}
          </div>
        </div>

        <div className="config-section">
          <label className="config-label">最大时长</label>
          <div className="param-row">
            <input
              type="number" className="input-sm"
              value={config.maxDuration || ''}
              placeholder="0 = 不限"
              onChange={e => updateConfig({ maxDuration: Number(e.target.value) || 0 })}
              disabled={isRecording}
            />
            <span className="param-unit">秒 (0=不限)</span>
          </div>
        </div>

        <div className="config-section">
          <label className="config-label">保存路径</label>
          <div className="param-row">
            <input
              type="text" className="input-path"
              value={config.outputDir}
              onChange={e => updateConfig({ outputDir: e.target.value })}
              readOnly={isRecording}
              placeholder="点击浏览选择"
            />
            <button
              className="btn btn-secondary"
              disabled={isRecording}
              onClick={selectOutputDir}
            >浏览</button>
          </div>
        </div>
      </div>

      {/* Error */}
      {error && (
        <div className="recorder-error">{error}</div>
      )}

      {/* Log Panel */}
      {isRecording && logs.length > 0 && (
        <div className="log-panel" style={{ gridColumn: '1 / -1' }}>
          <div className="log-header" onClick={() => {}}>
            <span className="log-header-title">录制日志 ({logs.length})</span>
          </div>
          <pre className="log-content">
            {logs.slice(-50).map((line, i) => (
              <div key={i} className={`log-line ${line.startsWith('[命令') || line.startsWith('[错误') || line.startsWith('[完成') ? 'log-line-highlight' : ''}`}>
                {line}
              </div>
            ))}
          </pre>
        </div>
      )}
    </div>
  )
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/RecorderPanel.tsx
git commit -m "feat(recorder): add RecorderPanel component with full UI"
```

---

## Task 10: App.tsx Tab 集成 + CSS 样式

**Files:**
- Modify: `frontend/src/App.tsx`
- Modify: `frontend/src/App.css`

- [ ] **Step 1: 修改 App.tsx 添加 Tab 切换**

在 `App.tsx` 中：

1. 新增 import：
```tsx
import RecorderPanel from './components/RecorderPanel'
```

2. 新增 state（在 `const [version, setVersion] = useState('')` 之后）：
```tsx
const [activeTab, setActiveTab] = useState<'extract' | 'record'>('extract')
```

3. 修改 toolbar，在 `toolbar-brand` 和 `toolbar-actions` 之间插入 tabs：
```tsx
<div className="toolbar-tabs">
  <button
    className={`toolbar-tab ${activeTab === 'extract' ? 'active' : ''}`}
    onClick={() => setActiveTab('extract')}
  >抽帧</button>
  <button
    className={`toolbar-tab ${activeTab === 'record' ? 'active' : ''}`}
    onClick={() => setActiveTab('record')}
  >录制</button>
</div>
```

4. 用条件渲染包裹现有内容和录制面板（在 header 之后、ActionBar 之前）：
```tsx
{activeTab === 'extract' ? (
  <>
    <TaskList ... />
    {selectedTask && <ConfigPanel ... />}
    {selectedTask && selectedTask.status !== 'pending' && (
      <div className="log-panel">...</div>
    )}
  </>
) : (
  <RecorderPanel />
)}
```

5. ActionBar 仅在 extract tab 显示：
```tsx
{activeTab === 'extract' && (
  <ActionBar ... />
)}
```

- [ ] **Step 2: 添加录制相关 CSS 到 App.css**

在 `App.css` 末尾追加：

```css
/* ── Toolbar Tabs ── */
.toolbar-tabs {
  display: flex;
  gap: 2px;
  background: var(--color-bg-tertiary);
  border-radius: var(--radius-sm);
  padding: 2px;
}

.toolbar-tab {
  padding: 4px 14px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  color: var(--color-text-muted);
  background: transparent;
  transition: all 0.15s ease;
}

.toolbar-tab:hover {
  color: var(--color-text-secondary);
}

.toolbar-tab.active {
  background: var(--color-bg-secondary);
  color: var(--color-text);
  box-shadow: 0 1px 3px rgba(0,0,0,0.2);
}

/* ── Recorder Panel ── */
.recorder-panel {
  flex: 1;
  display: grid;
  grid-template-columns: 1.2fr 0.8fr;
  gap: 0;
  overflow: hidden;
}

.recorder-left {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 16px;
  border-right: 1px solid var(--color-border-subtle);
}

.recorder-right {
  padding: 16px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

/* ── Preview ── */
.recorder-preview {
  flex: 1;
  background: #000;
  border-radius: var(--radius);
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 200px;
}

.recorder-preview img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.recorder-preview-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: var(--color-text-muted);
  font-size: 13px;
}

.preview-icon {
  font-size: 32px;
}

/* ── Recording Overlay ── */
.recorder-recording-overlay {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.recording-indicator {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #fff;
  font-size: 14px;
  font-weight: 500;
}

.recording-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #e53e3e;
  animation: pulse 1.5s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.3; }
}

.recording-time {
  color: #fff;
  font-size: 28px;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}

/* ── Controls ── */
.recorder-controls {
  display: flex;
  gap: 6px;
}

.recorder-btn-record {
  flex: 1;
}

/* ── Status Bar ── */
.recorder-status {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: var(--color-text-muted);
  font-variant-numeric: tabular-nums;
}

/* ── Error ── */
.recorder-error {
  color: var(--color-error);
  font-size: 12px;
  padding: 8px 16px;
  background: var(--color-error-dim);
  border-top: 1px solid var(--color-border-subtle);
  grid-column: 1 / -1;
}

/* ── Select ── */
.recorder-select {
  width: 100%;
  padding: 5px 8px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  font-size: 12px;
  background: var(--color-bg);
  color: var(--color-text);
  outline: none;
}

.recorder-select:focus {
  border-color: var(--color-primary);
}
```

- [ ] **Step 3: 确认前端编译通过**

Run: `cd /Users/wangjin/Downloads/video-extractor/frontend && npm run build`
Expected: 编译成功

- [ ] **Step 4: Commit**

```bash
git add frontend/src/App.tsx frontend/src/App.css
git commit -m "feat(recorder): integrate tab switching and recorder UI into App"
```

---

## Task 11: FFmpeg 编译配置更新

**Files:**
- Modify: `.github/workflows/build-ffmpeg.yml`

- [ ] **Step 1: 更新 macOS 构建步骤**

macOS 构建中，在 `./configure` 之前添加 libx264 安装步骤，并更新 configure 参数：

在 `Build FFmpeg` step 中，`./configure` 之前添加：
```bash
brew install x264
```

更新 macOS `./configure` 块，在现有参数之后追加：
```
--enable-avfoundation \
--enable-indev=avfoundation \
--enable-protocol=pipe \
--enable-muxer=mp4,matroska,image2pipe \
--enable-encoder=libx264,aac \
--enable-decoder=rawvideo \
--enable-libx264 \
--enable-gpl \
```

保留所有现有的 `--enable-*` 参数不变。

- [ ] **Step 2: 更新 Windows 构建步骤**

在 MSYS2 安装的包列表中追加 `mingw-w64-x86_64-x264`：
```yaml
install: >-
  mingw-w64-x86_64-gcc
  mingw-w64-x86_64-nasm
  mingw-w64-x86_64-x264
  make
  git
  diffutils
```

更新 Windows `./configure` 块，追加：
```
--enable-dshow \
--enable-indev=dshow \
--enable-protocol=pipe \
--enable-muxer=mp4,matroska,image2pipe \
--enable-encoder=libx264,aac \
--enable-decoder=rawvideo \
--enable-libx264 \
--enable-gpl \
```

注意：Windows 构建保持 `--extra-ldflags="-static"` 和 `--enable-w32threads` 不变。

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/build-ffmpeg.yml
git commit -m "build: update FFmpeg config with avfoundation/dshow/libx264 for recording"
```

---

## Task 12: 集成测试与最终验证

- [ ] **Step 1: 运行全部 Go 测试**

Run: `cd /Users/wangjin/Downloads/video-extractor && go test ./... -v`
Expected: 所有测试 PASS

- [ ] **Step 2: 运行前端构建**

Run: `cd /Users/wangjin/Downloads/video-extractor/frontend && npm run build`
Expected: 构建成功

- [ ] **Step 3: 启动应用手动验证**

Run: `cd /Users/wangjin/Downloads/video-extractor && wails3 dev`

验证清单：
- [ ] 工具栏显示「抽帧」和「录制」两个 Tab
- [ ] 点击「录制」Tab 切换到录制界面
- [ ] 摄像头列表正确显示
- [ ] 选择摄像头后预览画面出现
- [ ] 配置面板各项参数可正常调整
- [ ] 点击「开始录制」后预览消失，显示录制状态
- [ ] 录制中状态信息实时更新（时长/大小/帧率）
- [ ] 点击「停止」结束录制，文件保存正确
- [ ] 切回「抽帧」Tab，原有功能正常
- [ ] 日志面板显示 FFmpeg 输出

- [ ] **Step 4: 最终 Commit**

```bash
git add -A
git commit -m "feat: camera recording module complete"
```
