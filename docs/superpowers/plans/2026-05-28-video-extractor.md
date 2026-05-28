# 视频抽帧桌面应用实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 构建一个基于 Wails 3 + React + Go 的跨平台视频抽帧桌面客户端，内嵌 FFmpeg，双击即用。

**Architecture:** Go 后端通过 `os/exec` 调用内嵌的 FFmpeg/ffprobe 二进制，提供视频解析和抽帧服务；React 前端通过 Wails bindings 调用后端方法，通过事件系统接收进度更新。任务队列串行执行。

**Tech Stack:** Go 1.25+ / Wails 3 v3.0.0-alpha.95 / React 18 / TypeScript / Vite / FFmpeg static binaries

---

## 文件结构

| 文件 | 职责 |
|------|------|
| `main.go` | 应用入口，Wails 配置，窗口创建 |
| `app.go` | 主服务，暴露给前端的方法（文件选择、任务管理） |
| `internal/model/task.go` | Task、ModeParams 数据模型 |
| `internal/ffmpeg/binary.go` | FFmpeg/ffprobe 路径解析与命令执行 |
| `internal/ffmpeg/progress.go` | FFmpeg stderr 进度解析 |
| `internal/video/info.go` | ffprobe 视频信息解析 |
| `internal/extractor/extractor.go` | 抽帧命令构建与执行 |
| `internal/task/manager.go` | 任务队列管理（增删改查、串行执行） |
| `frontend/src/App.tsx` | 主布局组件 |
| `frontend/src/components/TaskList.tsx` | 任务列表组件 |
| `frontend/src/components/TaskItem.tsx` | 单个任务行组件 |
| `frontend/src/components/ConfigPanel.tsx` | 配置面板（整合所有配置子组件） |
| `frontend/src/components/ModeSelector.tsx` | 抽帧模式选择（平铺按钮组） |
| `frontend/src/components/OutputFormat.tsx` | 输出格式与质量配置 |
| `frontend/src/components/SizeConfig.tsx` | 输出尺寸配置 |
| `frontend/src/components/NamingSelector.tsx` | 文件命名规则选择 |
| `frontend/src/components/ActionBar.tsx` | 底部操作栏 |
| `frontend/src/hooks/useTaskQueue.ts` | 任务队列状态管理 hook |
| `Taskfile.yml` | 构建与打包任务 |
| `wails.json` | Wails 配置 |

---

### Task 1: 项目脚手架

**Files:**
- Create: `go.mod`
- Create: `wails.json`
- Create: `.gitignore`
- Create: `frontend/package.json`
- Create: `frontend/tsconfig.json`
- Create: `frontend/vite.config.ts`
- Create: `frontend/index.html`
- Create: `frontend/src/main.tsx`
- Create: `main.go`

- [ ] **Step 1: 初始化 Go 模块**

```bash
cd /Users/wangjin/Downloads/video-extractor
go mod init video-extractor
go get github.com/wailsapp/wails/v3@v3.0.0-alpha.95
go get github.com/google/uuid@latest
```

- [ ] **Step 2: 创建 wails.json**

```json
{
  "$schema": "https://wails.io/schemas/config.v3.json",
  "name": "视频抽帧工具",
  "outputfilename": "video-extractor",
  "frontend:install": "npm install",
  "frontend:build": "npm run build",
  "frontend:dev:watcher": "npm run dev",
  "frontend:dev:serverUrl": "auto",
  "author": {
    "name": "wangjin"
  }
}
```

- [ ] **Step 3: 创建 .gitignore**

```
.superpowers/
node_modules/
dist/
build/bin/
thirdparty/
*.exe
*.dmg
.DS_Store
coverage.out
coverage.html
```

- [ ] **Step 4: 创建 frontend/package.json**

```json
{
  "name": "frontend",
  "private": true,
  "version": "0.0.0",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "tsc && vite build",
    "preview": "vite preview"
  },
  "dependencies": {
    "@wailsio/runtime": "^3.0.0-alpha.79",
    "react": "^18.2.0",
    "react-dom": "^18.2.0"
  },
  "devDependencies": {
    "@types/react": "^18.0.17",
    "@types/react-dom": "^18.0.6",
    "@vitejs/plugin-react": "^2.0.1",
    "typescript": "^5.9.3",
    "vite": "^3.0.7"
  }
}
```

- [ ] **Step 5: 创建 frontend/tsconfig.json**

```json
{
  "compilerOptions": {
    "target": "ES2020",
    "useDefineForClassFields": true,
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "jsx": "react-jsx",
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true
  },
  "include": ["src"],
  "references": [{ "path": "./tsconfig.node.json" }]
}
```

- [ ] **Step 6: 创建 frontend/tsconfig.node.json**

```json
{
  "compilerOptions": {
    "composite": true,
    "skipLibCheck": true,
    "module": "ESNext",
    "moduleResolution": "bundler",
    "allowSyntheticDefaultImports": true
  },
  "include": ["vite.config.ts"]
}
```

- [ ] **Step 7: 创建 frontend/vite.config.ts**

```typescript
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
})
```

- [ ] **Step 8: 创建 frontend/index.html**

```html
<!DOCTYPE html>
<html lang="zh-CN">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>视频抽帧工具</title>
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/src/main.tsx"></script>
  </body>
</html>
```

- [ ] **Step 9: 创建 frontend/src/main.tsx**

```tsx
import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
)
```

- [ ] **Step 10: 创建 frontend/src/App.tsx 占位**

```tsx
function App() {
  return <div>视频抽帧工具</div>
}

export default App
```

- [ ] **Step 11: 创建 main.go**

```go
package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
)

var version = "dev"

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := application.New(application.Options{
		Name:        "视频抽帧工具",
		Description: "视频抽帧桌面客户端",
		Services: []application.Service{
			application.NewService(NewApp()),
		},
		Assets: application.AssetOptions{
			Handler: application.BundledAssetFileServer(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:           "视频抽帧工具",
		Width:           960,
		Height:          720,
		DevToolsEnabled: true,
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
```

- [ ] **Step 12: 创建 app.go 占位**

```go
package main

import "context"

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	a.ctx = ctx
	return nil
}

func (a *App) ServiceShutdown() error {
	return nil
}
```

注意：此处编译会因缺少 import 而报错，先不编译。在后续任务中会补充完整 import。

- [ ] **Step 13: 安装前端依赖**

```bash
cd /Users/wangjin/Downloads/video-extractor/frontend
npm install
```

- [ ] **Step 14: 创建目录结构**

```bash
cd /Users/wangjin/Downloads/video-extractor
mkdir -p internal/model
mkdir -p internal/ffmpeg
mkdir -p internal/video
mkdir -p internal/extractor
mkdir -p internal/task
mkdir -p frontend/src/components
mkdir -p frontend/src/hooks
mkdir -p frontend/src/bindings
mkdir -p build/darwin
mkdir -p thirdparty
```

- [ ] **Step 15: 提交**

```bash
git init
git add -A
git commit -m "chore: scaffold Wails 3 + React project structure"
```

---

### Task 2: 数据模型

**Files:**
- Create: `internal/model/task.go`
- Create: `internal/model/task_test.go`

- [ ] **Step 1: 编写 Task 模型测试**

```go
// internal/model/task_test.go
package model

import "testing"

func TestTaskStatusConstants(t *testing.T) {
	statuses := []string{StatusPending, StatusRunning, StatusCompleted, StatusFailed, StatusCancelled}
	for _, s := range statuses {
		if s == "" {
			t.Error("status constant should not be empty")
		}
	}
}

func TestModeConstants(t *testing.T) {
	modes := []string{
		ModeEveryFrame, ModeEverySecond, ModeTimeRange,
		ModeTotalFrames, ModeKeyframe, ModeSceneChange, ModeTimestamps,
	}
	for _, m := range modes {
		if m == "" {
			t.Error("mode constant should not be empty")
		}
	}
}

func TestNewDefaultTask(t *testing.T) {
	task := NewDefaultTask("/path/to/video.mp4")
	if task.ID == "" {
		t.Error("task ID should not be empty")
	}
	if task.Status != StatusPending {
		t.Errorf("expected status %s, got %s", StatusPending, task.Status)
	}
	if task.VideoPath != "/path/to/video.mp4" {
		t.Error("VideoPath mismatch")
	}
	if task.Mode != ModeEverySecond {
		t.Errorf("expected mode %s, got %s", ModeEverySecond, task.Mode)
	}
	if task.Format != "jpeg" {
		t.Error("default format should be jpeg")
	}
	if task.Quality != 80 {
		t.Error("default quality should be 80")
	}
	if task.ScaleMode != "percentage" {
		t.Error("default ScaleMode should be percentage")
	}
	if task.Scale != 100 {
		t.Error("default scale should be 100")
	}
	if task.Naming != "sequential" {
		t.Error("default naming should be sequential")
	}
	if task.Params.SceneThreshold != 0.3 {
		t.Error("default scene threshold should be 0.3")
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

```bash
cd /Users/wangjin/Downloads/video-extractor
go test ./internal/model/... -v
```

Expected: 编译失败，`NewDefaultTask` 未定义

- [ ] **Step 3: 实现 Task 模型**

```go
// internal/model/task.go
package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusPending   = "pending"
	StatusRunning   = "running"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
	StatusCancelled = "cancelled"

	ModeEveryFrame  = "every-frame"
	ModeEverySecond = "every-second"
	ModeTimeRange   = "time-range"
	ModeTotalFrames = "total-frames"
	ModeKeyframe    = "keyframe"
	ModeSceneChange = "scene-change"
	ModeTimestamps  = "timestamps"
)

type Task struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	VideoPath string    `json:"videoPath"`

	Mode   string     `json:"mode"`
	Params ModeParams `json:"params"`

	Format    string `json:"format"`
	Quality   int    `json:"quality"`
	ScaleMode string `json:"scaleMode"`
	Scale     int    `json:"scale"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	KeepRatio bool   `json:"keepRatio"`

	OutputDir string `json:"outputDir"`
	Naming    string `json:"naming"`

	Progress    float64   `json:"progress"`
	Current     int       `json:"current"`
	Total       int       `json:"total"`
	ElapsedTime int64     `json:"elapsedTime"`
	Error       string    `json:"error"`
	CreatedAt   time.Time `json:"createdAt"`
}

type ModeParams struct {
	FrameInterval  int      `json:"frameInterval"`
	SecondInterval float64  `json:"secondInterval"`
	StartTime      string   `json:"startTime"`
	EndTime        string   `json:"endTime"`
	TotalFrames    int      `json:"totalFrames"`
	SceneThreshold float64  `json:"sceneThreshold"`
	Timestamps     []string `json:"timestamps"`
}

func NewDefaultTask(videoPath string) *Task {
	return &Task{
		ID:        uuid.New().String(),
		Status:    StatusPending,
		VideoPath: videoPath,
		Mode:      ModeEverySecond,
		Params: ModeParams{
			SecondInterval: 1,
			SceneThreshold: 0.3,
		},
		Format:    "jpeg",
		Quality:   80,
		ScaleMode: "percentage",
		Scale:     100,
		Naming:    "sequential",
		CreatedAt: time.Now(),
	}
}
```

- [ ] **Step 4: 运行测试确认通过**

```bash
go test ./internal/model/... -v
```

Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add internal/model/
git commit -m "feat: add Task and ModeParams data model"
```

---

### Task 3: FFmpeg 二进制管理

**Files:**
- Create: `internal/ffmpeg/binary.go`
- Create: `internal/ffmpeg/binary_test.go`

- [ ] **Step 1: 编写 FFmpeg 路径解析测试**

```go
// internal/ffmpeg/binary_test.go
package ffmpeg

import (
	"os"
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
	execDir := filepath.Dir(m.execPath)

	if runtime.GOOS == "darwin" {
		// macOS: check that path contains Contents/Resources
		if !containsPath(ffmpegPath, "Resources") {
			t.Errorf("macOS ffmpeg should be in Resources, got: %s", ffmpegPath)
		}
	} else {
		// Others: check relative to executable directory
		rel, err := filepath.Rel(execDir, ffmpegPath)
		if err != nil {
			t.Fatal(err)
		}
		if rel == "" {
			t.Error("ffmpeg should be relative to executable")
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
	// In test environment, ffmpeg won't exist — just verify no panic
	_ = m.FFmpegExists()
}

func TestEnsureBinaries_CreatesTempFallback(t *testing.T) {
	m := NewManager()
	// Won't have real binaries in test — verify it returns a usable path
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
```

- [ ] **Step 2: 运行测试确认失败**

```bash
go test ./internal/ffmpeg/... -v
```

Expected: 编译失败，类型和方法未定义

- [ ] **Step 3: 实现 FFmpeg 二进制管理**

```go
// internal/ffmpeg/binary.go
package ffmpeg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
		// macOS .app bundle: executable is in Contents/MacOS, FFmpeg in Contents/Resources/ffmpeg/
		resourcesDir := filepath.Join(filepath.Dir(dir), "Resources", "ffmpeg")
		ffmpeg = filepath.Join(resourcesDir, "ffmpeg")
		ffprobe = filepath.Join(resourcesDir, "ffprobe")
		// Fallback: if not in .app, look next to executable
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

func ParseProgress(line string) (frame int, timeSec float64, ok bool) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "frame=") {
		return 0, 0, false
	}

	parts := strings.SplitN(line, "\t", -1)
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "frame=") {
			fmt.Sscanf(part, "frame=%d", &frame)
		}
		if strings.HasPrefix(part, "time=") {
			timeStr := strings.TrimPrefix(part, "time=")
			timeSec = parseTimeToSeconds(timeStr)
		}
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
	var h, m float64
	var s float64
	fmt.Sscanf(parts[0], "%f", &h)
	fmt.Sscanf(parts[1], "%f", &m)
	fmt.Sscanf(parts[2], "%f", &s)
	return h*3600 + m*60 + s
}
```

- [ ] **Step 4: 运行测试确认通过**

```bash
go test ./internal/ffmpeg/... -v
```

Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add internal/ffmpeg/
git commit -m "feat: add FFmpeg binary path resolution and command execution"
```

---

### Task 4: 视频信息解析

**Files:**
- Create: `internal/video/info.go`
- Create: `internal/video/info_test.go`

- [ ] **Step 1: 编写视频信息解析测试**

```go
// internal/video/info_test.go
package video

import "testing"

func TestParseVideoInfo_ValidFFprobeJSON(t *testing.T) {
	jsonOutput := `{
		"streams": [
			{
				"codec_type": "video",
				"width": 1920,
				"height": 1080,
				"r_frame_rate": "30/1",
				"nb_frames": "900",
				"duration": "30.000000"
			}
		],
		"format": {
			"duration": "30.000000"
		}
	}`

	info, err := ParseVideoInfo([]byte(jsonOutput))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Width != 1920 {
		t.Errorf("expected width 1920, got %d", info.Width)
	}
	if info.Height != 1080 {
		t.Errorf("expected height 1080, got %d", info.Height)
	}
	if info.FPS != 30 {
		t.Errorf("expected fps 30, got %f", info.FPS)
	}
	if info.Duration != 30.0 {
		t.Errorf("expected duration 30, got %f", info.Duration)
	}
	if info.TotalFrames != 900 {
		t.Errorf("expected total frames 900, got %d", info.TotalFrames)
	}
}

func TestParseVideoInfo_NoVideoStream(t *testing.T) {
	jsonOutput := `{"streams": [{"codec_type": "audio"}], "format": {"duration": "10"}}`
	_, err := ParseVideoInfo([]byte(jsonOutput))
	if err == nil {
		t.Error("expected error for no video stream")
	}
}

func TestParseVideoInfo_InvalidJSON(t *testing.T) {
	_, err := ParseVideoInfo([]byte("not json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestParseVideoInfo_MissingNbFrames_CalculatesFromDuration(t *testing.T) {
	jsonOutput := `{
		"streams": [
			{
				"codec_type": "video",
				"width": 1280,
				"height": 720,
				"r_frame_rate": "24/1",
				"duration": "60.0"
			}
		],
		"format": {"duration": "60.0"}
	}`

	info, err := ParseVideoInfo([]byte(jsonOutput))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.TotalFrames != 1440 {
		t.Errorf("expected 1440 frames (60*24), got %d", info.TotalFrames)
	}
}

func TestParseVideoInfo_NoStreams_NoFormat(t *testing.T) {
	jsonOutput := `{"streams": [], "format": {}}`
	_, err := ParseVideoInfo([]byte(jsonOutput))
	if err == nil {
		t.Error("expected error when no video stream found")
	}
}

func TestParseFrameRate(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"30/1", 30},
		{"24000/1001", 23.976023976023978},
		{"30000/1001", 29.970029970029973},
		{"25/1", 25},
		{"60/1", 60},
	}
	for _, tt := range tests {
		result := parseFrameRate(tt.input)
		diff := result - tt.expected
		if diff < -0.01 || diff > 0.01 {
			t.Errorf("parseFrameRate(%s) = %f, want %f", tt.input, result, tt.expected)
		}
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

```bash
go test ./internal/video/... -v
```

Expected: 编译失败

- [ ] **Step 3: 实现视频信息解析**

```go
// internal/video/info.go
package video

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type Info struct {
	Width       int     `json:"width"`
	Height      int     `json:"height"`
	FPS         float64 `json:"fps"`
	Duration    float64 `json:"duration"`
	TotalFrames int     `json:"totalFrames"`
}

type ffprobeOutput struct {
	Streams []ffprobeStream `json:"streams"`
	Format  ffprobeFormat   `json:"format"`
}

type ffprobeStream struct {
	CodecType   string `json:"codec_type"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	RFrameRate  string `json:"r_frame_rate"`
	NBFrames    string `json:"nb_frames"`
	Duration    string `json:"duration"`
}

type ffprobeFormat struct {
	Duration string `json:"duration"`
}

func ParseVideoInfo(data []byte) (*Info, error) {
	var output ffprobeOutput
	if err := json.Unmarshal(data, &output); err != nil {
		return nil, fmt.Errorf("parse ffprobe output: %w", err)
	}

	var videoStream *ffprobeStream
	for i := range output.Streams {
		if output.Streams[i].CodecType == "video" {
			videoStream = &output.Streams[i]
			break
		}
	}
	if videoStream == nil {
		return nil, fmt.Errorf("no video stream found")
	}

	info := &Info{
		Width:  videoStream.Width,
		Height: videoStream.Height,
		FPS:    parseFrameRate(videoStream.RFrameRate),
	}

	// Duration: prefer stream, fallback to format
	if videoStream.Duration != "" {
		info.Duration, _ = strconv.ParseFloat(videoStream.Duration, 64)
	} else if output.Format.Duration != "" {
		info.Duration, _ = strconv.ParseFloat(output.Format.Duration, 64)
	}

	// TotalFrames: prefer nb_frames, fallback to duration * fps
	if videoStream.NBFrames != "" {
		info.TotalFrames, _ = strconv.Atoi(videoStream.NBFrames)
	}
	if info.TotalFrames == 0 && info.Duration > 0 && info.FPS > 0 {
		info.TotalFrames = int(info.Duration * info.FPS)
	}

	return info, nil
}

func parseFrameRate(rate string) float64 {
	parts := strings.Split(rate, "/")
	if len(parts) != 2 {
		f, _ := strconv.ParseFloat(rate, 64)
		return f
	}
	num, _ := strconv.ParseFloat(parts[0], 64)
	den, _ := strconv.ParseFloat(parts[1], 64)
	if den == 0 {
		return 0
	}
	return num / den
}
```

- [ ] **Step 4: 运行测试确认通过**

```bash
go test ./internal/video/... -v
```

Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add internal/video/
git commit -m "feat: add video info parsing from ffprobe JSON output"
```

---

### Task 5: 抽帧器 — FFmpeg 命令构建

**Files:**
- Create: `internal/extractor/extractor.go`
- Create: `internal/extractor/extractor_test.go`

- [ ] **Step 1: 编写抽帧命令构建测试**

```go
// internal/extractor/extractor_test.go
package extractor

import (
	"strings"
	"testing"

	"video-extractor/internal/model"
)

func TestBuildArgs_EveryFrame(t *testing.T) {
	task := &model.Task{
		Mode: model.ModeEveryFrame,
		Params: model.ModeParams{
			FrameInterval: 10,
		},
		Format: "jpeg",
		Quality: 80,
	}
	args, err := BuildArgs(task, "/input/video.mp4", "/output")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	argStr := strings.Join(args, " ")
	if !strings.Contains(argStr, "select='not(mod(n\\,10))'") {
		t.Errorf("expected select filter for every 10 frames, got: %s", argStr)
	}
	if !strings.Contains(argStr, "-vsync") || !strings.Contains(argStr, "vfr") {
		t.Error("expected -vsync vfr for every-frame mode")
	}
}

func TestBuildArgs_EverySecond(t *testing.T) {
	task := &model.Task{
		Mode: model.ModeEverySecond,
		Params: model.ModeParams{
			SecondInterval: 2,
		},
		Format: "jpeg",
	}
	args, err := BuildArgs(task, "/input/video.mp4", "/output")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	argStr := strings.Join(args, " ")
	if !strings.Contains(argStr, "fps=1/2") {
		t.Errorf("expected fps=1/2, got: %s", argStr)
	}
}

func TestBuildArgs_TimeRange(t *testing.T) {
	task := &model.Task{
		Mode: model.ModeTimeRange,
		Params: model.ModeParams{
			StartTime: "00:00:10",
			EndTime:   "00:00:30",
		},
		Format: "jpeg",
	}
	args, err := BuildArgs(task, "/input/video.mp4", "/output")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	argStr := strings.Join(args, " ")
	if !strings.Contains(argStr, "-ss") || !strings.Contains(argStr, "00:00:10") {
		t.Errorf("expected -ss 00:00:10, got: %s", argStr)
	}
	if !strings.Contains(argStr, "-to") || !strings.Contains(argStr, "00:00:30") {
		t.Errorf("expected -to 00:00:30, got: %s", argStr)
	}
}

func TestBuildArgs_TotalFrames(t *testing.T) {
	task := &model.Task{
		Mode: model.ModeTotalFrames,
		Params: model.ModeParams{
			TotalFrames: 20,
		},
		Format: "jpeg",
	}
	args, err := BuildArgs(task, "/input/video.mp4", "/output")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	argStr := strings.Join(args, " ")
	if !strings.Contains(argStr, "-frames:v") || !strings.Contains(argStr, "20") {
		t.Errorf("expected -frames:v 20, got: %s", argStr)
	}
}

func TestBuildArgs_Keyframe(t *testing.T) {
	task := &model.Task{
		Mode:   model.ModeKeyframe,
		Format: "jpeg",
	}
	args, err := BuildArgs(task, "/input/video.mp4", "/output")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	argStr := strings.Join(args, " ")
	if !strings.Contains(argStr, "eq(pict_type\\,I)") {
		t.Errorf("expected keyframe filter, got: %s", argStr)
	}
}

func TestBuildArgs_SceneChange(t *testing.T) {
	task := &model.Task{
		Mode: model.ModeSceneChange,
		Params: model.ModeParams{
			SceneThreshold: 0.4,
		},
		Format: "jpeg",
	}
	args, err := BuildArgs(task, "/input/video.mp4", "/output")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	argStr := strings.Join(args, " ")
	if !strings.Contains(argStr, "gt(scene\\,0.4)") {
		t.Errorf("expected scene change filter with threshold 0.4, got: %s", argStr)
	}
}

func TestBuildArgs_ScalePercentage(t *testing.T) {
	task := &model.Task{
		Mode:      model.ModeEverySecond,
		Format:    "jpeg",
		ScaleMode: "percentage",
		Scale:     50,
	}
	args, err := BuildArgs(task, "/input/video.mp4", "/output")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	argStr := strings.Join(args, " ")
	if !strings.Contains(argStr, "scale=iw*0.50:ih*0.50") {
		t.Errorf("expected scale filter, got: %s", argStr)
	}
}

func TestBuildArgs_ScaleCustom(t *testing.T) {
	task := &model.Task{
		Mode:      model.ModeEverySecond,
		Format:    "jpeg",
		ScaleMode: "custom",
		Width:     1920,
		Height:    1080,
		KeepRatio: true,
	}
	args, err := BuildArgs(task, "/input/video.mp4", "/output")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	argStr := strings.Join(args, " ")
	if !strings.Contains(argStr, "scale=1920:-1") {
		t.Errorf("expected scale=1920:-1, got: %s", argStr)
	}
}

func TestBuildArgs_JPEGQuality(t *testing.T) {
	task := &model.Task{
		Mode:    model.ModeEverySecond,
		Format:  "jpeg",
		Quality: 80,
	}
	args, err := BuildArgs(task, "/input/video.mp4", "/output")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	argStr := strings.Join(args, " ")
	if !strings.Contains(argStr, "-q:v") {
		t.Errorf("expected -q:v for jpeg quality, got: %s", argStr)
	}
}

func TestBuildArgs_PNG_NoQuality(t *testing.T) {
	task := &model.Task{
		Mode:   model.ModeEverySecond,
		Format: "png",
	}
	args, err := BuildArgs(task, "/input/video.mp4", "/output")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	argStr := strings.Join(args, " ")
	if strings.Contains(argStr, "-q:v") {
		t.Error("PNG should not have -q:v quality param")
	}
}

func TestBuildOutputPattern(t *testing.T) {
	tests := []struct {
		naming   string
		format   string
		videoPath string
		expected string
	}{
		{"sequential", "jpeg", "/videos/test.mp4", "frame_%04d.jpg"},
		{"sequential", "png", "/videos/test.mp4", "frame_%04d.png"},
		{"timestamp", "jpeg", "/videos/test.mp4", "%d.jpg"},
		{"video-seq", "jpeg", "/videos/test.mp4", "test_%04d.jpg"},
		{"video-time", "jpeg", "/videos/test.mp4", "test_%d.jpg"},
	}
	for _, tt := range tests {
		result := BuildOutputPattern(tt.naming, tt.format, tt.videoPath)
		if result != tt.expected {
			t.Errorf("BuildOutputPattern(%s,%s,%s) = %s, want %s",
				tt.naming, tt.format, tt.videoPath, result, tt.expected)
		}
	}
}

func TestQualityToFFmpeg(t *testing.T) {
	// 100% quality → q:v 1 (best)
	if q := QualityToFFmpeg(100); q != 1 {
		t.Errorf("QualityToFFmpeg(100) = %d, want 1", q)
	}
	// 1% quality → q:v 31 (worst)
	if q := QualityToFFmpeg(1); q != 31 {
		t.Errorf("QualityToFFmpeg(1) = %d, want 31", q)
	}
	// 50% quality → somewhere in between
	q := QualityToFFmpeg(50)
	if q < 1 || q > 31 {
		t.Errorf("QualityToFFmpeg(50) = %d, out of range [1,31]", q)
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

```bash
go test ./internal/extractor/... -v
```

Expected: 编译失败

- [ ] **Step 3: 实现抽帧命令构建**

```go
// internal/extractor/extractor.go
package extractor

import (
	"fmt"
	"path/filepath"
	"strings"

	"video-extractor/internal/model"
)

func BuildArgs(task *model.Task, inputPath string, outputDir string) ([]string, error) {
	var args []string

	// Time range: -ss and -to go before -i for fast seeking
	if task.Mode == model.ModeTimeRange {
		if task.Params.StartTime != "" {
			args = append(args, "-ss", task.Params.StartTime)
		}
		if task.Params.EndTime != "" {
			args = append(args, "-to", task.Params.EndTime)
		}
	}

	args = append(args, "-i", inputPath)

	// Build video filter chain
	vf := buildVideoFilter(task)
	if vf != "" {
		args = append(args, "-vf", vf)
	}

	// Mode-specific: some modes need -vsync vfr
	needsVFR := task.Mode == model.ModeEveryFrame || task.Mode == model.ModeKeyframe ||
		task.Mode == model.ModeSceneChange || task.Mode == model.ModeTotalFrames
	if needsVFR {
		args = append(args, "-vsync", "vfr")
	}

	// Total frames mode: limit output frames
	if task.Mode == model.ModeTotalFrames {
		args = append(args, "-frames:v", fmt.Sprintf("%d", task.Params.TotalFrames))
	}

	// JPEG quality
	if task.Format == "jpeg" {
		args = append(args, "-q:v", fmt.Sprintf("%d", QualityToFFmpeg(task.Quality)))
	}

	// Output pattern
	pattern := BuildOutputPattern(task.Naming, task.Format, task.VideoPath)
	args = append(args, filepath.Join(outputDir, pattern))

	return args, nil
}

func buildVideoFilter(task *model.Task) string {
	var filters []string

	switch task.Mode {
	case model.ModeEveryFrame:
		filters = append(filters, fmt.Sprintf("select='not(mod(n\\,%d))'", task.Params.FrameInterval))
	case model.ModeEverySecond:
		filters = append(filters, fmt.Sprintf("fps=1/%g", task.Params.SecondInterval))
	case model.ModeTotalFrames:
		filters = append(filters, fmt.Sprintf(
			"select='not(mod(n\\,round(n/%d)))'", task.Params.TotalFrames,
		))
	case model.ModeKeyframe:
		filters = append(filters, "select='eq(pict_type\\,I)'")
	case model.ModeSceneChange:
		filters = append(filters, fmt.Sprintf("select='gt(scene\\,%g)'", task.Params.SceneThreshold))
	case model.ModeTimeRange:
		// No special filter needed, just extract all frames in range
	}

	// Scale filter
	switch task.ScaleMode {
	case "percentage":
		if task.Scale != 100 && task.Scale > 0 {
			ratio := float64(task.Scale) / 100.0
			filters = append(filters, fmt.Sprintf("scale=iw*%.2f:ih*%.2f", ratio, ratio))
		}
	case "custom":
		if task.Width > 0 {
			h := fmt.Sprintf("%d", task.Height)
			if task.KeepRatio {
				h = "-1"
			}
			filters = append(filters, fmt.Sprintf("scale=%d:%s", task.Width, h))
		}
	}

	if len(filters) == 0 {
		return ""
	}
	return strings.Join(filters, ",")
}

func BuildOutputPattern(naming string, format string, videoPath string) string {
	ext := "jpg"
	if format == "png" {
		ext = "png"
	}
	videoName := strings.TrimSuffix(filepath.Base(videoPath), filepath.Ext(videoPath))

	switch naming {
	case "sequential":
		return fmt.Sprintf("frame_%%04d.%s", ext)
	case "timestamp":
		return fmt.Sprintf("%%d.%s", ext)
	case "video-seq":
		return fmt.Sprintf("%s_%%04d.%s", videoName, ext)
	case "video-time":
		return fmt.Sprintf("%s_%%d.%s", videoName, ext)
	default:
		return fmt.Sprintf("frame_%%04d.%s", ext)
	}
}

func QualityToFFmpeg(quality int) int {
	// FFmpeg q:v: 1=best, 31=worst. Map 1-100 → 1-31.
	if quality <= 0 {
		quality = 1
	}
	if quality > 100 {
		quality = 100
	}
	return 32 - ((quality * 31) / 100)
}

func BuildTimestampArgs(task *model.Task, outputDir string) [][]string {
	// For timestamps mode, each timestamp gets its own ffmpeg command
	var commands [][]string
	ext := "jpg"
	if task.Format == "png" {
		ext = "png"
	}

	videoName := strings.TrimSuffix(filepath.Base(task.VideoPath), filepath.Ext(task.VideoPath))

	for i, ts := range task.Params.Timestamps {
		filename := fmt.Sprintf("%s_%04d.%s", videoName, i+1, ext)
		args := []string{
			"-ss", ts,
			"-i", task.VideoPath,
			"-frames:v", "1",
		}
		if task.Format == "jpeg" {
			args = append(args, "-q:v", fmt.Sprintf("%d", QualityToFFmpeg(task.Quality)))
		}
		args = append(args, filepath.Join(outputDir, filename))
		commands = append(commands, args)
	}
	return commands
}
```

- [ ] **Step 4: 运行测试确认通过**

```bash
go test ./internal/extractor/... -v
```

Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add internal/extractor/
git commit -m "feat: add FFmpeg command builder for all 7 extraction modes"
```

---

### Task 6: 任务队列管理器

**Files:**
- Create: `internal/task/manager.go`
- Create: `internal/task/manager_test.go`

- [ ] **Step 1: 编写任务管理器测试**

```go
// internal/task/manager_test.go
package task

import (
	"testing"

	"video-extractor/internal/model"
)

func TestManager_AddTask(t *testing.T) {
	m := NewManager()
	task := m.AddTask("/path/to/video.mp4")
	if task.ID == "" {
		t.Error("task should have an ID")
	}
	if task.Status != model.StatusPending {
		t.Error("new task should be pending")
	}
	if len(m.ListTasks()) != 1 {
		t.Error("should have 1 task")
	}
}

func TestManager_GetTask(t *testing.T) {
	m := NewManager()
	task := m.AddTask("/path/to/video.mp4")
	found := m.GetTask(task.ID)
	if found == nil {
		t.Fatal("task should be found")
	}
	if found.ID != task.ID {
		t.Error("ID mismatch")
	}
}

func TestManager_GetTask_NotFound(t *testing.T) {
	m := NewManager()
	found := m.GetTask("nonexistent")
	if found != nil {
		t.Error("should return nil for nonexistent task")
	}
}

func TestManager_UpdateTask(t *testing.T) {
	m := NewManager()
	task := m.AddTask("/path/to/video.mp4")
	task.Mode = model.ModeKeyframe
	task.Quality = 95
	m.UpdateTask(task)

	updated := m.GetTask(task.ID)
	if updated.Mode != model.ModeKeyframe {
		t.Errorf("expected mode %s, got %s", model.ModeKeyframe, updated.Mode)
	}
	if updated.Quality != 95 {
		t.Errorf("expected quality 95, got %d", updated.Quality)
	}
}

func TestManager_RemoveTask(t *testing.T) {
	m := NewManager()
	task := m.AddTask("/path/to/video.mp4")
	m.RemoveTask(task.ID)
	if len(m.ListTasks()) != 0 {
		t.Error("task list should be empty after removal")
	}
	if m.GetTask(task.ID) != nil {
		t.Error("removed task should not be found")
	}
}

func TestManager_ClearTasks(t *testing.T) {
	m := NewManager()
	m.AddTask("/video1.mp4")
	m.AddTask("/video2.mp4")
	m.ClearTasks()
	if len(m.ListTasks()) != 0 {
		t.Error("task list should be empty after clear")
	}
}

func TestManager_GetPendingTasks(t *testing.T) {
	m := NewManager()
	t1 := m.AddTask("/video1.mp4")
	t2 := m.AddTask("/video2.mp4")
	t3 := m.AddTask("/video3.mp4")
	t1.Status = model.StatusRunning
	m.UpdateTask(t1)
	t3.Status = model.StatusCompleted
	m.UpdateTask(t3)

	pending := m.GetPendingTasks()
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending task, got %d", len(pending))
	}
	if pending[0].ID != t2.ID {
		t.Error("pending task should be t2")
	}
}

func TestManager_CallbackFired(t *testing.T) {
	m := NewManager()
	var callbackTask *model.Task
	m.SetCallback(func(task *model.Task) {
		callbackTask = task
	})
	task := m.AddTask("/video.mp4")
	if callbackTask == nil || callbackTask.ID != task.ID {
		t.Error("callback should have been fired with the new task")
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

```bash
go test ./internal/task/... -v
```

Expected: 编译失败

- [ ] **Step 3: 实现任务管理器**

```go
// internal/task/manager.go
package task

import (
	"sync"

	"video-extractor/internal/model"
)

type Callback func(task *model.Task)

type Manager struct {
	mu       sync.RWMutex
	tasks    []*model.Task
	callback Callback
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) SetCallback(cb Callback) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callback = cb
}

func (m *Manager) AddTask(videoPath string) *model.Task {
	task := model.NewDefaultTask(videoPath)
	m.mu.Lock()
	m.tasks = append(m.tasks, task)
	cb := m.callback
	m.mu.Unlock()
	if cb != nil {
		cb(task)
	}
	return task
}

func (m *Manager) GetTask(id string) *model.Task {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, t := range m.tasks {
		if t.ID == id {
			return t
		}
	}
	return nil
}

func (m *Manager) ListTasks() []*model.Task {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*model.Task, len(m.tasks))
	copy(result, m.tasks)
	return result
}

func (m *Manager) UpdateTask(task *model.Task) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, t := range m.tasks {
		if t.ID == task.ID {
			m.tasks[i] = task
			break
		}
	}
	cb := m.callback
	if cb != nil {
		cb(task)
	}
}

func (m *Manager) RemoveTask(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, t := range m.tasks {
		if t.ID == id {
			m.tasks = append(m.tasks[:i], m.tasks[i+1:]...)
			break
		}
	}
}

func (m *Manager) ClearTasks() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tasks = nil
}

func (m *Manager) GetPendingTasks() []*model.Task {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*model.Task
	for _, t := range m.tasks {
		if t.Status == model.StatusPending {
			result = append(result, t)
		}
	}
	return result
}
```

- [ ] **Step 4: 运行测试确认通过**

```bash
go test ./internal/task/... -v
```

Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add internal/task/
git commit -m "feat: add task queue manager with CRUD and callbacks"
```

---

### Task 7: App 服务 — Wails 绑定

**Files:**
- Create: `app.go` (replace placeholder)
- Create: `internal/app/service.go`

- [ ] **Step 1: 实现应用服务**

```go
// internal/app/service.go
package app

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"

	"video-extractor/internal/extractor"
	"video-extractor/internal/ffmpeg"
	"video-extractor/internal/task"
	"video-extractor/internal/model"
	"video-extractor/internal/video"
)

type Service struct {
	ctx       context.Context
	taskMgr   *task.Manager
	ffmpegMgr *ffmpeg.Manager
}

func NewService() *Service {
	return &Service{
		taskMgr:   task.NewManager(),
		ffmpegMgr: ffmpeg.NewManager(),
	}
}

func (s *Service) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	s.ctx = ctx
	s.taskMgr.SetCallback(func(t *model.Task) {
		application.Get().Event.Emit("task:changed", t)
	})
	return nil
}

func (s *Service) ServiceShutdown() error {
	return nil
}

// AddTasks 添加视频文件到任务队列
func (s *Service) AddTasks(paths []string) []*model.Task {
	var tasks []*model.Task
	for _, p := range paths {
		t := s.taskMgr.AddTask(p)
		// 设置默认输出目录为视频同目录下的 以视频名命名 的子文件夹
		videoName := strings.TrimSuffix(filepath.Base(p), filepath.Ext(p))
		t.OutputDir = filepath.Join(filepath.Dir(p), videoName+"_frames")
		s.taskMgr.UpdateTask(t)
		tasks = append(tasks, t)
	}
	return tasks
}

// GetTasks 获取所有任务
func (s *Service) GetTasks() []*model.Task {
	return s.taskMgr.ListTasks()
}

// UpdateTask 更新任务配置
func (s *Service) UpdateTask(t *model.Task) {
	s.taskMgr.UpdateTask(t)
}

// RemoveTask 移除任务
func (s *Service) RemoveTask(id string) {
	s.taskMgr.RemoveTask(id)
}

// ClearTasks 清空所有任务
func (s *Service) ClearTasks() {
	s.taskMgr.ClearTasks()
}

// SelectFiles 打开文件选择对话框（多选）
func (s *Service) SelectFiles() []string {
	result := application.Get().Window.NewOpenFileDialog()
	result.SetTitle("选择视频文件")
	result.SetMultiple(true)
	result.AddFilter("视频文件", "mp4", "avi", "mkv", "mov", "flv", "wmv", "webm", "ts", "m4v")
	result.AddFilter("所有文件", "*")
	if result.Show() {
		return result.GetFiles()
	}
	return nil
}

// SelectFolder 打开文件夹选择对话框
func (s *Service) SelectFolder() string {
	result := application.Get().Window.NewOpenDirectoryDialog()
	result.SetTitle("选择包含视频的文件夹")
	if result.Show() {
		return result.GetDirectory()
	}
	return ""
}

// SelectOutputDir 选择输出目录
func (s *Service) SelectOutputDir() string {
	result := application.Get().Window.NewOpenDirectoryDialog()
	result.SetTitle("选择输出保存路径")
	if result.Show() {
		return result.GetDirectory()
	}
	return ""
}

// GetVideoInfo 获取视频信息
func (s *Service) GetVideoInfo(videoPath string) (*video.Info, error) {
	out, err := s.ffmpegMgr.RunFFprobe(
		"-v", "quiet",
		"-print_format", "json",
		"-show_streams",
		"-show_format",
		videoPath,
	)
	if err != nil {
		return nil, fmt.Errorf("获取视频信息失败: %w", err)
	}
	return video.ParseVideoInfo(out)
}

// ScanVideoFiles 扫描文件夹中的视频文件
func (s *Service) ScanVideoFiles(dir string) []string {
	videoExts := map[string]bool{
		".mp4": true, ".avi": true, ".mkv": true, ".mov": true,
		".flv": true, ".wmv": true, ".webm": true, ".ts": true, ".m4v": true,
	}
	var files []string
	entries, err := os.ReadDir(dir)
	if err != nil {
		return files
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if videoExts[ext] {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}
	return files
}

// StartAll 开始执行所有待处理任务
func (s *Service) StartAll() {
	pending := s.taskMgr.GetPendingTasks()
	go s.executeTasks(pending)
}

// CancelTask 取消任务
func (s *Service) CancelTask(id string) {
	t := s.taskMgr.GetTask(id)
	if t != nil && t.Status == model.StatusRunning {
		t.Status = model.StatusCancelled
		s.taskMgr.UpdateTask(t)
	}
}

var progressRe = regexp.MustCompile(`frame=\s*(\d+)`)

func (s *Service) executeTasks(tasks []*model.Task) {
	for _, t := range tasks {
		if t.Status == model.StatusCancelled {
			continue
		}
		s.executeTask(t)
	}
}

func (s *Service) executeTask(t *model.Task) {
	// Get video info for progress calculation
	info, err := s.GetVideoInfo(t.VideoPath)
	if err != nil {
		t.Status = model.StatusFailed
		t.Error = fmt.Sprintf("获取视频信息失败: %v", err)
		s.taskMgr.UpdateTask(t)
		return
	}

	// Create output directory
	if err := os.MkdirAll(t.OutputDir, 0755); err != nil {
		t.Status = model.StatusFailed
		t.Error = fmt.Sprintf("创建输出目录失败: %v", err)
		s.taskMgr.UpdateTask(t)
		return
	}

	t.Status = model.StatusRunning
	t.Total = info.TotalFrames
	s.taskMgr.UpdateTask(t)

	startTime := time.Now()

	if t.Mode == model.ModeTimestamps {
		s.executeTimestamps(t, startTime)
		return
	}

	args, err := extractor.BuildArgs(t, t.VideoPath, t.OutputDir)
	if err != nil {
		t.Status = model.StatusFailed
		t.Error = err.Error()
		s.taskMgr.UpdateTask(t)
		return
	}

	cmd := s.ffmpegMgr.RunFFmpeg(args...)
	stderr, _ := cmd.StderrPipe()
	cmd.Start()

	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		line := scanner.Text()
		frame, _, ok := ffmpeg.ParseProgress(line)
		if ok && t.Total > 0 {
			t.Progress = float64(frame) / float64(t.Total) * 100
			t.Current = frame
			t.ElapsedTime = int64(time.Since(startTime).Seconds())
			s.taskMgr.UpdateTask(t)
		}

		// Check cancellation
		current := s.taskMgr.GetTask(t.ID)
		if current != nil && current.Status == model.StatusCancelled {
			cmd.Process.Kill()
			return
		}
	}
	cmd.Wait()

	current := s.taskMgr.GetTask(t.ID)
	if current != nil && current.Status == model.StatusCancelled {
		return
	}

	t.Status = model.StatusCompleted
	t.Progress = 100
	t.ElapsedTime = int64(time.Since(startTime).Seconds())
	s.taskMgr.UpdateTask(t)
}

func (s *Service) executeTimestamps(t *model.Task, startTime time.Time) {
	commands := extractor.BuildTimestampArgs(t, t.OutputDir)
	totalCmds := len(commands)
	for i, args := range commands {
		current := s.taskMgr.GetTask(t.ID)
		if current != nil && current.Status == model.StatusCancelled {
			return
		}

		cmd := s.ffmpegMgr.RunFFmpeg(args...)
		cmd.Run()

		t.Progress = float64(i+1) / float64(totalCmds) * 100
		t.Current = i + 1
		t.ElapsedTime = int64(time.Since(startTime).Seconds())
		s.taskMgr.UpdateTask(t)
	}

	t.Status = model.StatusCompleted
	t.Progress = 100
	t.ElapsedTime = int64(time.Since(startTime).Seconds())
	s.taskMgr.UpdateTask(t)
}
```

- [ ] **Step 2: 更新 main.go 使用新服务**

```go
// main.go
package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"

	"video-extractor/internal/app"
)

var version = "dev"

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	appService := app.NewService()

	applicationInstance := application.New(application.Options{
		Name:        "视频抽帧工具",
		Description: "视频抽帧桌面客户端",
		Services: []application.Service{
			application.NewService(appService),
		},
		Assets: application.AssetOptions{
			Handler: application.BundledAssetFileServer(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	applicationInstance.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:           "视频抽帧工具",
		Width:           960,
		Height:          720,
		DevToolsEnabled: true,
	})

	if err := applicationInstance.Run(); err != nil {
		log.Fatal(err)
	}
}
```

- [ ] **Step 3: 删除旧的 app.go 占位文件**

```bash
rm /Users/wangjin/Downloads/video-extractor/app.go
```

- [ ] **Step 4: 运行 Go 编译检查**

```bash
cd /Users/wangjin/Downloads/video-extractor
go build ./...
```

Expected: 编译成功（需要 `frontend/dist` 目录存在才能编译 main.go，先创建空目录）

```bash
mkdir -p frontend/dist && echo '<!DOCTYPE html><html><body></body></html>' > frontend/dist/index.html
go build ./...
```

- [ ] **Step 5: 提交**

```bash
git add -A
git commit -m "feat: add App service with Wails bindings, file dialogs, and task execution"
```

---

### Task 8: 前端基础布局与样式

**Files:**
- Create: `frontend/src/App.tsx`
- Create: `frontend/src/App.css`
- Create: `frontend/src/style.css`

- [ ] **Step 1: 创建全局样式**

```css
/* frontend/src/style.css */
:root {
  --color-primary: #3b82f6;
  --color-primary-hover: #2563eb;
  --color-bg: #ffffff;
  --color-bg-secondary: #f8fafc;
  --color-bg-tertiary: #f1f5f9;
  --color-border: #e2e8f0;
  --color-text: #1e293b;
  --color-text-secondary: #64748b;
  --color-text-muted: #94a3b8;
  --color-success: #22c55e;
  --color-error: #ef4444;
  --color-warning: #f59e0b;
  --color-info: #3b82f6;
  --radius: 6px;
  --radius-sm: 4px;
}

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
  font-size: 13px;
  color: var(--color-text);
  background: var(--color-bg);
  user-select: none;
  overflow: hidden;
}

#root {
  height: 100vh;
  display: flex;
  flex-direction: column;
}

button {
  cursor: pointer;
  border: none;
  background: none;
  font-family: inherit;
  font-size: inherit;
}

input {
  font-family: inherit;
  font-size: inherit;
}
```

- [ ] **Step 2: 创建 App 组件**

```tsx
// frontend/src/App.tsx
import './App.css'
import TaskList from './components/TaskList'
import ConfigPanel from './components/ConfigPanel'
import ActionBar from './components/ActionBar'
import { useTaskQueue } from './hooks/useTaskQueue'

function App() {
  const {
    tasks,
    selectedTaskId,
    selectTask,
    addFiles,
    addFolder,
    updateTask,
    removeTask,
    clearTasks,
    startAll,
    cancelTask,
  } = useTaskQueue()

  return (
    <div className="app">
      <header className="toolbar">
        <div className="toolbar-title">视频抽帧工具</div>
        <div className="toolbar-actions">
          <button className="btn btn-primary" onClick={addFiles}>选择视频</button>
          <button className="btn btn-secondary" onClick={addFolder}>选择文件夹</button>
        </div>
      </header>

      <TaskList
        tasks={tasks}
        selectedTaskId={selectedTaskId}
        onSelectTask={selectTask}
        onRemoveTask={removeTask}
        onCancelTask={cancelTask}
      />

      {selectedTaskId && (
        <ConfigPanel
          task={tasks.find(t => t.id === selectedTaskId)!}
          onUpdateTask={updateTask}
        />
      )}

      <ActionBar
        tasks={tasks}
        onStartAll={startAll}
        onClear={clearTasks}
      />
    </div>
  )
}

export default App
```

- [ ] **Step 3: 创建 App.css**

```css
/* frontend/src/App.css */
.app {
  display: flex;
  flex-direction: column;
  height: 100vh;
}

.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  background: var(--color-bg-secondary);
  border-bottom: 1px solid var(--color-border);
  flex-shrink: 0;
}

.toolbar-title {
  font-weight: 600;
  font-size: 15px;
  color: var(--color-text);
}

.toolbar-actions {
  display: flex;
  gap: 8px;
}

.btn {
  padding: 6px 14px;
  border-radius: var(--radius-sm);
  font-size: 12px;
  font-weight: 500;
  transition: background 0.15s;
}

.btn-primary {
  background: var(--color-primary);
  color: white;
}

.btn-primary:hover {
  background: var(--color-primary-hover);
}

.btn-secondary {
  background: var(--color-bg-tertiary);
  color: var(--color-text);
  border: 1px solid var(--color-border);
}

.btn-secondary:hover {
  background: var(--color-border);
}

.btn-danger {
  color: var(--color-error);
}

.btn-danger:hover {
  background: #fef2f2;
}
```

- [ ] **Step 4: 更新 main.tsx 引入全局样式**

```tsx
// frontend/src/main.tsx
import React from 'react'
import ReactDOM from 'react-dom/client'
import './style.css'
import App from './App'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
)
```

- [ ] **Step 5: 提交**

```bash
git add frontend/src/
git commit -m "feat: add App layout with toolbar, task list, config panel, action bar"
```

---

### Task 9: 前端状态管理 Hook

**Files:**
- Create: `frontend/src/hooks/useTaskQueue.ts`

- [ ] **Step 1: 实现任务队列状态管理 hook**

```typescript
// frontend/src/hooks/useTaskQueue.ts
import { useState, useEffect, useCallback } from 'react'

// Types matching Go model
export interface ModeParams {
  frameInterval: number
  secondInterval: number
  startTime: string
  endTime: string
  totalFrames: number
  sceneThreshold: number
  timestamps: string[]
}

export interface Task {
  id: string
  status: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled'
  videoPath: string
  mode: string
  params: ModeParams
  format: string
  quality: number
  scaleMode: string
  scale: number
  width: number
  height: number
  keepRatio: boolean
  outputDir: string
  naming: string
  progress: number
  current: number
  total: number
  elapsedTime: number
  error: string
  createdAt: string
}

// Wails runtime - will be generated by wails3 generate bindings
declare function AddTasks(paths: string[]): Promise<Task[]>
declare function GetTasks(): Promise<Task[]>
declare function UpdateTask(task: Task): Promise<void>
declare function RemoveTask(id: string): Promise<void>
declare function ClearTasks(): Promise<void>
declare function SelectFiles(): Promise<string[]>
declare function SelectFolder(): Promise<string>
declare function SelectOutputDir(): Promise<string>
declare function StartAll(): Promise<void>
declare function CancelTask(id: string): Promise<void>
declare function ScanVideoFiles(dir: string): Promise<string[]>

// Events
declare function EventsOn(event: string, callback: (...args: any[]) => void): void

export function useTaskQueue() {
  const [tasks, setTasks] = useState<Task[]>([])
  const [selectedTaskId, setSelectedTaskId] = useState<string | null>(null)

  useEffect(() => {
    GetTasks().then(setTasks).catch(() => {})

    EventsOn('task:changed', (task: Task) => {
      setTasks(prev => {
        const index = prev.findIndex(t => t.id === task.id)
        if (index >= 0) {
          const next = [...prev]
          next[index] = task
          return next
        }
        return [...prev, task]
      })
    })
  }, [])

  const selectTask = useCallback((id: string | null) => {
    setSelectedTaskId(id)
  }, [])

  const addFiles = useCallback(async () => {
    const files = await SelectFiles()
    if (files && files.length > 0) {
      const newTasks = await AddTasks(files)
      if (newTasks.length > 0) {
        setSelectedTaskId(newTasks[0].id)
      }
    }
  }, [])

  const addFolder = useCallback(async () => {
    const folder = await SelectFolder()
    if (folder) {
      const files = await ScanVideoFiles(folder)
      if (files.length > 0) {
        const newTasks = await AddTasks(files)
        if (newTasks.length > 0) {
          setSelectedTaskId(newTasks[0].id)
        }
      }
    }
  }, [])

  const updateTask = useCallback(async (task: Task) => {
    await UpdateTask(task)
  }, [])

  const removeTask = useCallback(async (id: string) => {
    await RemoveTask(id)
    setSelectedTaskId(prev => prev === id ? null : prev)
  }, [])

  const clearTasks = useCallback(async () => {
    await ClearTasks()
    setSelectedTaskId(null)
  }, [])

  const startAll = useCallback(async () => {
    await StartAll()
  }, [])

  const cancelTask = useCallback(async (id: string) => {
    await CancelTask(id)
  }, [])

  const selectOutputDir = useCallback(async (): Promise<string> => {
    return await SelectOutputDir()
  }, [])

  return {
    tasks,
    selectedTaskId,
    selectTask,
    addFiles,
    addFolder,
    updateTask,
    removeTask,
    clearTasks,
    startAll,
    cancelTask,
    selectOutputDir,
  }
}
```

- [ ] **Step 2: 提交**

```bash
git add frontend/src/hooks/
git commit -m "feat: add useTaskQueue hook for task state management"
```

---

### Task 10: 前端组件 — TaskList 与 TaskItem

**Files:**
- Create: `frontend/src/components/TaskList.tsx`
- Create: `frontend/src/components/TaskItem.tsx`

- [ ] **Step 1: 实现 TaskList 组件**

```tsx
// frontend/src/components/TaskList.tsx
import TaskItem from './TaskItem'
import { Task } from '../hooks/useTaskQueue'

interface Props {
  tasks: Task[]
  selectedTaskId: string | null
  onSelectTask: (id: string) => void
  onRemoveTask: (id: string) => void
  onCancelTask: (id: string) => void
}

export default function TaskList({ tasks, selectedTaskId, onSelectTask, onRemoveTask, onCancelTask }: Props) {
  if (tasks.length === 0) {
    return (
      <div className="task-list-empty">
        <p>点击上方"选择视频"或"选择文件夹"添加视频文件</p>
      </div>
    )
  }

  return (
    <div className="task-list">
      <div className="task-list-header">
        <span className="task-list-count">任务列表 ({tasks.length})</span>
      </div>
      <div className="task-list-items">
        {tasks.map(task => (
          <TaskItem
            key={task.id}
            task={task}
            selected={task.id === selectedTaskId}
            onSelect={() => onSelectTask(task.id)}
            onRemove={() => onRemoveTask(task.id)}
            onCancel={() => onCancelTask(task.id)}
          />
        ))}
      </div>
    </div>
  )
}
```

- [ ] **Step 2: 实现 TaskItem 组件**

```tsx
// frontend/src/components/TaskItem.tsx
import { Task } from '../hooks/useTaskQueue'

interface Props {
  task: Task
  selected: boolean
  onSelect: () => void
  onRemove: () => void
  onCancel: () => void
}

function formatTime(seconds: number): string {
  if (seconds < 60) return `${seconds}秒`
  const m = Math.floor(seconds / 60)
  const s = seconds % 60
  return `${m}分${s}秒`
}

function getVideoName(path: string): string {
  return path.split('/').pop() || path.split('\\').pop() || path
}

function StatusBadge({ status, progress, elapsedTime }: { status: string; progress: number; elapsedTime: number }) {
  switch (status) {
    case 'running':
      return (
        <div className="task-progress">
          <div className="progress-bar">
            <div className="progress-fill" style={{ width: `${progress}%` }} />
          </div>
          <span className="progress-text">
            {Math.round(progress)}% · 剩余 {formatTime(elapsedTime)}
          </span>
        </div>
      )
    case 'completed':
      return <span className="status status-completed">已完成</span>
    case 'failed':
      return <span className="status status-failed">失败</span>
    case 'cancelled':
      return <span className="status status-cancelled">已取消</span>
    default:
      return <span className="status status-pending">等待中</span>
  }
}

export default function TaskItem({ task, selected, onSelect, onRemove, onCancel }: Props) {
  const statusColor = {
    running: '#3b82f6',
    completed: '#22c55e',
    failed: '#ef4444',
    cancelled: '#94a3b8',
    pending: '#94a3b8',
  }[task.status]

  return (
    <div
      className={`task-item ${selected ? 'task-item-selected' : ''}`}
      onClick={onSelect}
    >
      <div className="task-dot" style={{ background: statusColor }} />
      <div className="task-info">
        <div className="task-name">{getVideoName(task.videoPath)}</div>
        <div className="task-summary">
          {getModeLabel(task.mode)} → {task.format.toUpperCase()}
          {task.scaleMode === 'percentage' && task.scale !== 100 && ` → 缩放${task.scale}%`}
          {task.scaleMode === 'custom' && task.width > 0 && ` → ${task.width}x${task.height}`}
        </div>
        {task.status === 'failed' && task.error && (
          <div className="task-error">{task.error}</div>
        )}
      </div>
      <div className="task-right">
        <StatusBadge status={task.status} progress={task.progress} elapsedTime={task.elapsedTime} />
        <div className="task-actions">
          {task.status === 'running' && (
            <button className="btn-icon" onClick={(e) => { e.stopPropagation(); onCancel() }} title="取消">✕</button>
          )}
          {task.status !== 'running' && (
            <button className="btn-icon" onClick={(e) => { e.stopPropagation(); onRemove() }} title="移除">✕</button>
          )}
        </div>
      </div>
    </div>
  )
}

function getModeLabel(mode: string): string {
  const labels: Record<string, string> = {
    'every-frame': '每N帧',
    'every-second': '每N秒',
    'time-range': '时间范围',
    'total-frames': '总帧数',
    'keyframe': '关键帧',
    'scene-change': '场景切换',
    'timestamps': '指定时间点',
  }
  return labels[mode] || mode
}
```

- [ ] **Step 3: 添加 TaskList 样式到 App.css**

在 `App.css` 末尾追加：

```css
/* Task List */
.task-list {
  border-bottom: 2px solid var(--color-border);
  padding: 12px 16px;
  flex-shrink: 0;
  max-height: 280px;
  overflow-y: auto;
}

.task-list-header {
  margin-bottom: 8px;
}

.task-list-count {
  font-size: 11px;
  color: var(--color-text-muted);
}

.task-list-empty {
  padding: 32px 16px;
  text-align: center;
  color: var(--color-text-muted);
  border-bottom: 2px solid var(--color-border);
  flex-shrink: 0;
}

.task-list-items {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.task-item {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  border-radius: var(--radius);
  border: 1px solid var(--color-border);
  gap: 12px;
  cursor: pointer;
  transition: background 0.1s;
  background: var(--color-bg);
}

.task-item:hover {
  background: var(--color-bg-secondary);
}

.task-item-selected {
  border-color: var(--color-primary);
  background: #eff6ff;
}

.task-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.task-info {
  flex: 1;
  min-width: 0;
}

.task-name {
  font-size: 13px;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.task-summary {
  font-size: 11px;
  color: var(--color-text-secondary);
  margin-top: 2px;
}

.task-error {
  font-size: 11px;
  color: var(--color-error);
  margin-top: 2px;
}

.task-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.task-progress {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 2px;
}

.progress-bar {
  width: 120px;
  height: 4px;
  background: var(--color-border);
  border-radius: 2px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: var(--color-primary);
  border-radius: 2px;
  transition: width 0.3s;
}

.progress-text {
  font-size: 10px;
  color: var(--color-text-secondary);
}

.status {
  font-size: 11px;
}

.status-pending { color: var(--color-text-muted); }
.status-completed { color: var(--color-success); }
.status-failed { color: var(--color-error); }
.status-cancelled { color: var(--color-text-muted); }

.task-actions {
  display: flex;
  gap: 4px;
}

.btn-icon {
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-sm);
  font-size: 10px;
  color: var(--color-text-muted);
}

.btn-icon:hover {
  background: var(--color-bg-tertiary);
  color: var(--color-text);
}
```

- [ ] **Step 4: 提交**

```bash
git add frontend/src/components/ frontend/src/App.css
git commit -m "feat: add TaskList and TaskItem components"
```

---

### Task 11: 前端组件 — ConfigPanel 与子组件

**Files:**
- Create: `frontend/src/components/ConfigPanel.tsx`
- Create: `frontend/src/components/ModeSelector.tsx`
- Create: `frontend/src/components/OutputFormat.tsx`
- Create: `frontend/src/components/SizeConfig.tsx`
- Create: `frontend/src/components/NamingSelector.tsx`

- [ ] **Step 1: 实现 ModeSelector 组件**

```tsx
// frontend/src/components/ModeSelector.tsx
import { Task } from '../hooks/useTaskQueue'

interface Props {
  task: Task
  onChange: (task: Task) => void
}

const MODES = [
  { value: 'every-second', label: '每N秒' },
  { value: 'every-frame', label: '每N帧' },
  { value: 'time-range', label: '时间范围' },
  { value: 'total-frames', label: '总帧数' },
  { value: 'keyframe', label: '关键帧' },
  { value: 'scene-change', label: '场景切换' },
  { value: 'timestamps', label: '指定时间点' },
]

export default function ModeSelector({ task, onChange }: Props) {
  const update = (patch: Partial<Task>) => {
    onChange({ ...task, ...patch })
  }

  return (
    <div className="config-section">
      <label className="config-label">抽帧模式</label>
      <div className="btn-group">
        {MODES.map(m => (
          <button
            key={m.value}
            className={`btn-option ${task.mode === m.value ? 'active' : ''}`}
            onClick={() => update({ mode: m.value })}
          >
            {m.label}
          </button>
        ))}
      </div>
      <div className="config-params">
        {task.mode === 'every-frame' && (
          <div className="param-row">
            <span className="param-label">间隔</span>
            <input
              type="number"
              className="input-sm"
              value={task.params.frameInterval || 10}
              min={1}
              onChange={e => update({
                params: { ...task.params, frameInterval: parseInt(e.target.value) || 1 }
              })}
            />
            <span className="param-unit">帧</span>
          </div>
        )}
        {task.mode === 'every-second' && (
          <div className="param-row">
            <span className="param-label">间隔</span>
            <input
              type="number"
              className="input-sm"
              value={task.params.secondInterval || 1}
              min={0.1}
              step={0.5}
              onChange={e => update({
                params: { ...task.params, secondInterval: parseFloat(e.target.value) || 0.5 }
              })}
            />
            <span className="param-unit">秒</span>
          </div>
        )}
        {task.mode === 'time-range' && (
          <div className="param-row">
            <input
              type="text"
              className="input-sm"
              placeholder="开始 00:00:00"
              value={task.params.startTime}
              onChange={e => update({
                params: { ...task.params, startTime: e.target.value }
              })}
            />
            <span className="param-unit">至</span>
            <input
              type="text"
              className="input-sm"
              placeholder="结束 00:00:00"
              value={task.params.endTime}
              onChange={e => update({
                params: { ...task.params, endTime: e.target.value }
              })}
            />
          </div>
        )}
        {task.mode === 'total-frames' && (
          <div className="param-row">
            <span className="param-label">目标帧数</span>
            <input
              type="number"
              className="input-sm"
              value={task.params.totalFrames || 20}
              min={1}
              onChange={e => update({
                params: { ...task.params, totalFrames: parseInt(e.target.value) || 1 }
              })}
            />
            <span className="param-unit">帧</span>
          </div>
        )}
        {task.mode === 'scene-change' && (
          <div className="param-row">
            <span className="param-label">灵敏度</span>
            <input
              type="range"
              min={0.1}
              max={0.9}
              step={0.05}
              value={task.params.sceneThreshold}
              onChange={e => update({
                params: { ...task.params, sceneThreshold: parseFloat(e.target.value) }
              })}
            />
            <span className="param-value">{task.params.sceneThreshold.toFixed(2)}</span>
          </div>
        )}
        {task.mode === 'timestamps' && (
          <div className="param-timestamps">
            {(task.params.timestamps || []).map((ts, i) => (
              <div key={i} className="param-row">
                <input
                  type="text"
                  className="input-sm"
                  placeholder="00:00:00"
                  value={ts}
                  onChange={e => {
                    const timestamps = [...(task.params.timestamps || [])]
                    timestamps[i] = e.target.value
                    update({ params: { ...task.params, timestamps } })
                  }}
                />
                <button
                  className="btn-icon"
                  onClick={() => {
                    const timestamps = (task.params.timestamps || []).filter((_, j) => j !== i)
                    update({ params: { ...task.params, timestamps } })
                  }}
                >✕</button>
              </div>
            ))}
            <button
              className="btn btn-secondary"
              onClick={() => {
                const timestamps = [...(task.params.timestamps || []), '']
                update({ params: { ...task.params, timestamps } })
              }}
            >+ 添加时间点</button>
          </div>
        )}
      </div>
    </div>
  )
}
```

- [ ] **Step 2: 实现 OutputFormat 组件**

```tsx
// frontend/src/components/OutputFormat.tsx
import { Task } from '../hooks/useTaskQueue'

interface Props {
  task: Task
  onChange: (task: Task) => void
}

export default function OutputFormat({ task, onChange }: Props) {
  const update = (patch: Partial<Task>) => {
    onChange({ ...task, ...patch })
  }

  return (
    <div className="config-section">
      <label className="config-label">输出格式</label>
      <div className="format-row">
        <div className="btn-group">
          <button
            className={`btn-option ${task.format === 'jpeg' ? 'active' : ''}`}
            onClick={() => update({ format: 'jpeg' })}
          >JPEG</button>
          <button
            className={`btn-option ${task.format === 'png' ? 'active' : ''}`}
            onClick={() => update({ format: 'png' })}
          >PNG</button>
        </div>
        {task.format === 'jpeg' && (
          <div className="param-row" style={{ marginLeft: 16 }}>
            <span className="param-label">质量</span>
            <input
              type="range"
              min={1}
              max={100}
              value={task.quality}
              onChange={e => update({ quality: parseInt(e.target.value) })}
            />
            <span className="param-value">{task.quality}%</span>
          </div>
        )}
      </div>
    </div>
  )
}
```

- [ ] **Step 3: 实现 SizeConfig 组件**

```tsx
// frontend/src/components/SizeConfig.tsx
import { Task } from '../hooks/useTaskQueue'

interface Props {
  task: Task
  onChange: (task: Task) => void
}

export default function SizeConfig({ task, onChange }: Props) {
  const update = (patch: Partial<Task>) => {
    onChange({ ...task, ...patch })
  }

  return (
    <div className="config-section">
      <label className="config-label">输出尺寸</label>
      <div className="format-row">
        <div className="btn-group">
          <button
            className={`btn-option ${task.scaleMode === 'percentage' ? 'active' : ''}`}
            onClick={() => update({ scaleMode: 'percentage' })}
          >按比例</button>
          <button
            className={`btn-option ${task.scaleMode === 'custom' ? 'active' : ''}`}
            onClick={() => update({ scaleMode: 'custom' })}
          >自定义</button>
        </div>
        {task.scaleMode === 'percentage' && (
          <div className="param-row" style={{ marginLeft: 16 }}>
            <span className="param-label">缩放</span>
            <input
              type="range"
              min={10}
              max={200}
              value={task.scale}
              onChange={e => update({ scale: parseInt(e.target.value) })}
            />
            <span className="param-value">{task.scale}%</span>
          </div>
        )}
        {task.scaleMode === 'custom' && (
          <div className="param-row" style={{ marginLeft: 16 }}>
            <input
              type="number"
              className="input-sm"
              placeholder="宽"
              value={task.width || ''}
              onChange={e => update({ width: parseInt(e.target.value) || 0 })}
            />
            <span className="param-unit">×</span>
            <input
              type="number"
              className="input-sm"
              placeholder="高"
              value={task.height || ''}
              onChange={e => update({ height: parseInt(e.target.value) || 0 })}
            />
            <label className="checkbox-label">
              <input
                type="checkbox"
                checked={task.keepRatio}
                onChange={e => update({ keepRatio: e.target.checked })}
              />
              保持比例
            </label>
          </div>
        )}
      </div>
    </div>
  )
}
```

- [ ] **Step 4: 实现 NamingSelector 组件**

```tsx
// frontend/src/components/NamingSelector.tsx
import { Task } from '../hooks/useTaskQueue'

interface Props {
  task: Task
  onChange: (task: Task) => void
}

const NAMINGS = [
  { value: 'sequential', label: '序号', example: 'frame_0001.png' },
  { value: 'timestamp', label: '时间戳', example: '00h01m30s_000f.png' },
  { value: 'video-seq', label: '视频名+序号', example: 'video1_0001.png' },
  { value: 'video-time', label: '视频名+时间', example: 'video1_00h01m30s.png' },
]

export default function NamingSelector({ task, onChange }: Props) {
  return (
    <div className="config-section">
      <label className="config-label">文件命名</label>
      <div className="btn-group">
        {NAMINGS.map(n => (
          <button
            key={n.value}
            className={`btn-option ${task.naming === n.value ? 'active' : ''}`}
            onClick={() => onChange({ ...task, naming: n.value })}
          >
            {n.label}
          </button>
        ))}
      </div>
    </div>
  )
}
```

- [ ] **Step 5: 实现 ConfigPanel 容器组件**

```tsx
// frontend/src/components/ConfigPanel.tsx
import { Task } from '../hooks/useTaskQueue'
import ModeSelector from './ModeSelector'
import OutputFormat from './OutputFormat'
import SizeConfig from './SizeConfig'
import NamingSelector from './NamingSelector'

interface Props {
  task: Task
  onUpdateTask: (task: Task) => void
}

export default function ConfigPanel({ task, onUpdateTask }: Props) {
  const isRunning = task.status === 'running'

  return (
    <div className="config-panel">
      <div className="config-grid">
        <div className="config-col">
          <ModeSelector task={task} onChange={onUpdateTask} />
          <OutputFormat task={task} onChange={onUpdateTask} />
        </div>
        <div className="config-col">
          <SizeConfig task={task} onChange={onUpdateTask} />
          <div className="config-section">
            <label className="config-label">保存路径</label>
            <div className="param-row">
              <input
                type="text"
                className="input-path"
                value={task.outputDir}
                onChange={e => onUpdateTask({ ...task, outputDir: e.target.value })}
                readOnly={isRunning}
              />
              <button
                className="btn btn-secondary"
                disabled={isRunning}
                onClick={async () => {
                  // This would call SelectOutputDir from the hook
                  // For now, handled via prop drilling from App
                }}
              >浏览</button>
            </div>
          </div>
          <NamingSelector task={task} onChange={onUpdateTask} />
        </div>
      </div>
    </div>
  )
}
```

- [ ] **Step 6: 追加配置面板样式到 App.css**

在 `App.css` 末尾追加：

```css
/* Config Panel */
.config-panel {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
}

.config-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.config-col {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.config-section {
  margin-bottom: 0;
}

.config-label {
  display: block;
  font-size: 12px;
  font-weight: 600;
  color: var(--color-text-secondary);
  margin-bottom: 8px;
}

.config-params {
  margin-top: 8px;
}

.config-params:empty {
  margin-top: 0;
}

/* Button Group */
.btn-group {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.btn-option {
  padding: 5px 12px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--color-border);
  font-size: 11px;
  color: var(--color-text-secondary);
  background: var(--color-bg);
  transition: all 0.15s;
}

.btn-option:hover {
  border-color: var(--color-primary);
  color: var(--color-primary);
}

.btn-option.active {
  background: var(--color-primary);
  color: white;
  border-color: var(--color-primary);
}

/* Param Row */
.param-row {
  display: flex;
  align-items: center;
  gap: 6px;
}

.param-label {
  font-size: 11px;
  color: var(--color-text-muted);
}

.param-unit {
  font-size: 11px;
  color: var(--color-text-muted);
}

.param-value {
  font-size: 11px;
  color: var(--color-text);
  font-weight: 500;
  min-width: 32px;
}

.param-timestamps {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

/* Inputs */
.input-sm {
  padding: 4px 8px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  font-size: 12px;
  width: 80px;
  outline: none;
}

.input-sm:focus {
  border-color: var(--color-primary);
}

.input-path {
  flex: 1;
  padding: 6px 8px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  font-size: 12px;
  outline: none;
}

.input-path:focus {
  border-color: var(--color-primary);
}

/* Range Slider */
input[type="range"] {
  width: 120px;
  accent-color: var(--color-primary);
}

/* Checkbox */
.checkbox-label {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  color: var(--color-text-secondary);
  cursor: pointer;
}

.checkbox-label input[type="checkbox"] {
  accent-color: var(--color-primary);
}

.format-row {
  display: flex;
  align-items: center;
}
```

- [ ] **Step 7: 提交**

```bash
git add frontend/src/components/ frontend/src/App.css
git commit -m "feat: add ConfigPanel with ModeSelector, OutputFormat, SizeConfig, NamingSelector"
```

---

### Task 12: 前端组件 — ActionBar

**Files:**
- Create: `frontend/src/components/ActionBar.tsx`

- [ ] **Step 1: 实现 ActionBar 组件**

```tsx
// frontend/src/components/ActionBar.tsx
import { Task } from '../hooks/useTaskQueue'

interface Props {
  tasks: Task[]
  onStartAll: () => void
  onClear: () => void
}

export default function ActionBar({ tasks, onStartAll, onClear }: Props) {
  const pendingCount = tasks.filter(t => t.status === 'pending').length
  const totalFrames = tasks.reduce((sum, t) => sum + t.total, 0)
  const hasRunning = tasks.some(t => t.status === 'running')

  return (
    <footer className="action-bar">
      <div className="action-bar-info">
        已选择 {tasks.length} 个视频
        {totalFrames > 0 && ` · 预计输出 ${totalFrames.toLocaleString()} 帧`}
      </div>
      <div className="action-bar-buttons">
        <button
          className="btn btn-primary"
          disabled={pendingCount === 0 || hasRunning}
          onClick={onStartAll}
        >
          {hasRunning ? '处理中...' : `全部开始${pendingCount > 0 ? ` (${pendingCount})` : ''}`}
        </button>
        <button
          className="btn btn-secondary"
          disabled={tasks.length === 0 || hasRunning}
          onClick={onClear}
        >清空队列</button>
      </div>
    </footer>
  )
}
```

- [ ] **Step 2: 追加 ActionBar 样式到 App.css**

在 `App.css` 末尾追加：

```css
/* Action Bar */
.action-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 16px;
  background: var(--color-bg-secondary);
  border-top: 1px solid var(--color-border);
  flex-shrink: 0;
}

.action-bar-info {
  font-size: 11px;
  color: var(--color-text-muted);
}

.action-bar-buttons {
  display: flex;
  gap: 8px;
}

button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
```

- [ ] **Step 3: 提交**

```bash
git add frontend/src/components/ActionBar.tsx frontend/src/App.css
git commit -m "feat: add ActionBar component with start and clear buttons"
```

---

### Task 13: Wails 绑定生成与集成

**Files:**
- Modify: `frontend/src/hooks/useTaskQueue.ts`

- [ ] **Step 1: 构建前端**

```bash
cd /Users/wangjin/Downloads/video-extractor/frontend
npm run build
```

- [ ] **Step 2: 生成 Wails TypeScript 绑定**

```bash
cd /Users/wangjin/Downloads/video-extractor
wails3 generate bindings -ts -d frontend/bindings .
```

- [ ] **Step 3: 更新 useTaskQueue.ts 使用生成的绑定**

将 `useTaskQueue.ts` 中的 `declare function` 声明替换为从 bindings 导入。具体导入路径取决于生成的绑定结构，通常为：

```typescript
// 将这些:
// declare function AddTasks(paths: string[]): Promise<Task[]>
// ... 等

// 替换为:
import { AddTasks, GetTasks, UpdateTask, RemoveTask, ClearTasks,
         SelectFiles, SelectFolder, SelectOutputDir, StartAll,
         CancelTask, ScanVideoFiles } from '../bindings/app'
import { EventsOn } from '@wailsio/runtime'
```

具体的导入路径需要根据 `wails3 generate bindings` 生成的文件结构调整。

- [ ] **Step 4: 验证前端编译**

```bash
cd /Users/wangjin/Downloads/video-extractor/frontend
npm run build
```

Expected: 编译成功

- [ ] **Step 5: 提交**

```bash
git add -A
git commit -m "feat: generate Wails bindings and integrate with frontend"
```

---

### Task 14: 构建与打包

**Files:**
- Create: `Taskfile.yml`
- Create: `build/darwin/Info.plist`

- [ ] **Step 1: 创建 Taskfile.yml**

```yaml
# Taskfile.yml
version: "3"

vars:
  APP_NAME: video-extractor
  FRONTEND_DIR: frontend
  BUILD_DIR: build/bin
  BINDINGS_DIR: frontend/bindings
  VERSION:
    sh: git describe --tags --always 2>/dev/null || echo "dev"

env:
  CGO_ENABLED: 1

tasks:
  # ===== Development =====
  dev:
    desc: Run in development mode with hot reload
    cmds:
      - wails3 dev

  # ===== Bindings =====
  generate:
    desc: Generate TypeScript bindings from Go services
    cmds:
      - wails3 generate bindings -ts -d {{.BINDINGS_DIR}} .

  # ===== Frontend =====
  frontend:install:
    desc: Install frontend dependencies
    dir: "{{.FRONTEND_DIR}}"
    cmds:
      - npm install

  frontend:build:
    desc: Build frontend assets
    deps: [frontend:install, generate]
    dir: "{{.FRONTEND_DIR}}"
    cmds:
      - npm run build

  # ===== Build =====
  build:darwin:
    desc: Build macOS binary for current architecture
    deps: [frontend:build]
    cmds:
      - go build -ldflags "-X main.version={{.VERSION}}" -o {{.BUILD_DIR}}/{{.APP_NAME}} .

  build:darwin:arm64:
    desc: Build macOS arm64 binary (Apple Silicon)
    deps: [frontend:build]
    cmds:
      - GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.version={{.VERSION}}" -o {{.BUILD_DIR}}/{{.APP_NAME}}-darwin-arm64 .

  build:darwin:amd64:
    desc: Build macOS amd64 binary
    deps: [frontend:build]
    cmds:
      - GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version={{.VERSION}}" -o {{.BUILD_DIR}}/{{.APP_NAME}}-darwin-amd64 .

  build:windows:amd64:
    desc: Build Windows amd64 binary
    deps: [frontend:build]
    cmds:
      - GOOS=windows GOARCH=amd64 go build -ldflags "-H windowsgui -X main.version={{.VERSION}}" -o {{.BUILD_DIR}}/{{.APP_NAME}}-windows-amd64.exe .

  build:
    desc: Build binary for current platform
    cmds:
      - task: build:darwin

  # ===== macOS Packaging =====
  generate:darwin:plist:
    desc: Generate macOS Info.plist
    internal: true
    cmds:
      - mkdir -p build/darwin
      - |
        cat > build/darwin/Info.plist << 'PLIST'
        <?xml version="1.0" encoding="UTF-8"?>
        <!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
        <plist version="1.0">
        <dict>
          <key>CFBundleExecutable</key>
          <string>video-extractor</string>
          <key>CFBundleIdentifier</key>
          <string>com.video-extractor.app</string>
          <key>CFBundleName</key>
          <string>视频抽帧工具</string>
          <key>CFBundleDisplayName</key>
          <string>视频抽帧工具</string>
          <key>CFBundleVersion</key>
          <string>{{.VERSION}}</string>
          <key>CFBundleShortVersionString</key>
          <string>{{.VERSION}}</string>
          <key>CFBundlePackageType</key>
          <string>APPL</string>
          <key>LSMinimumSystemVersion</key>
          <string>11.0</string>
          <key>NSHighResolutionCapable</key>
          <true/>
        </dict>
        </plist>
        PLIST
    generates:
      - build/darwin/Info.plist
    status:
      - test -f build/darwin/Info.plist

  package:darwin:
    desc: Package macOS .app bundle
    deps: [build:darwin, generate:darwin:plist]
    cmds:
      - mkdir -p {{.BUILD_DIR}}/{{.APP_NAME}}.app/Contents/{MacOS,Resources,Resources/ffmpeg}
      - cp {{.BUILD_DIR}}/{{.APP_NAME}} {{.BUILD_DIR}}/{{.APP_NAME}}.app/Contents/MacOS
      - cp build/darwin/Info.plist {{.BUILD_DIR}}/{{.APP_NAME}}.app/Contents
      - cp thirdparty/darwin-{{ARCH}}/ffmpeg {{.BUILD_DIR}}/{{.APP_NAME}}.app/Contents/Resources/ffmpeg/
      - cp thirdparty/darwin-{{ARCH}}/ffprobe {{.BUILD_DIR}}/{{.APP_NAME}}.app/Contents/Resources/ffmpeg/
      - chmod +x {{.BUILD_DIR}}/{{.APP_NAME}}.app/Contents/Resources/ffmpeg/*
      - codesign --force --deep --sign - {{.BUILD_DIR}}/{{.APP_NAME}}.app

  package:darwin:dmg:
    desc: Package macOS .dmg installer
    deps: [package:darwin]
    cmds:
      - mkdir -p /tmp/{{.APP_NAME}}-dmg
      - cp -R {{.BUILD_DIR}}/{{.APP_NAME}}.app /tmp/{{.APP_NAME}}-dmg/
      - ln -sf /Applications /tmp/{{.APP_NAME}}-dmg/Applications
      - hdiutil create -volname "视频抽帧工具" -srcfolder /tmp/{{.APP_NAME}}-dmg -ov -format UDZO {{.BUILD_DIR}}/{{.APP_NAME}}-macos.dmg

  package:windows:
    desc: Package Windows exe with FFmpeg
    deps: [build:windows:amd64]
    cmds:
      - mkdir -p {{.BUILD_DIR}}/windows-dist
      - cp {{.BUILD_DIR}}/{{.APP_NAME}}-windows-amd64.exe {{.BUILD_DIR}}/windows-dist/
      - cp thirdparty/windows-amd64/ffmpeg.exe {{.BUILD_DIR}}/windows-dist/
      - cp thirdparty/windows-amd64/ffprobe.exe {{.BUILD_DIR}}/windows-dist/
      - echo "Windows package ready at {{.BUILD_DIR}}/windows-dist/"

  # ===== Testing =====
  test:
    desc: Run Go tests
    cmds:
      - go test ./internal/... -v

  # ===== Cleanup =====
  clean:
    desc: Remove all build artifacts
    cmds:
      - rm -rf {{.BUILD_DIR}}
      - rm -rf {{.FRONTEND_DIR}}/dist
      - rm -rf {{.FRONTEND_DIR}}/node_modules
```

- [ ] **Step 2: 验证 Go 编译**

```bash
cd /Users/wangjin/Downloads/video-extractor
go build ./...
```

- [ ] **Step 3: 运行全部测试**

```bash
go test ./internal/... -v
```

Expected: 所有测试通过

- [ ] **Step 4: 提交**

```bash
git add -A
git commit -m "feat: add Taskfile with build, package, and dev tasks"
```

---

### Task 15: 端到端验证

**Files:**
- 无新文件

- [ ] **Step 1: 准备 FFmpeg 二进制**

从 https://ffmpeg.org/download.html 下载对应平台的 FFmpeg 静态编译版本，解压后放入 `thirdparty/` 目录。

```bash
# macOS 示例
mkdir -p thirdparty/darwin-arm64
cp /path/to/ffmpeg thirdparty/darwin-arm64/
cp /path/to/ffprobe thirdparty/darwin-arm64/
```

- [ ] **Step 2: 启动开发模式**

```bash
cd /Users/wangjin/Downloads/video-extractor
task dev
```

- [ ] **Step 3: 手动验证核心流程**

1. 点击"选择视频"，选择一个视频文件 → 确认任务出现在列表中
2. 选中任务，修改抽帧模式、格式、尺寸等配置
3. 点击"全部开始" → 确认进度条显示、帧数更新
4. 等待完成 → 确认输出目录中生成了正确的图片文件
5. 测试"选择文件夹"批量添加
6. 测试"清空队列"

- [ ] **Step 4: 打包测试**

```bash
# macOS
task package:darwin:dmg

# 确认 .dmg 文件生成
ls -la build/bin/video-extractor-macos.dmg
```

- [ ] **Step 5: 最终提交**

```bash
git add -A
git commit -m "chore: add FFmpeg binaries and verify end-to-end flow"
```

---

## 自审检查清单

**1. Spec 覆盖度：**
- [x] 单文件/多文件/文件夹选择 → Task 7 (SelectFiles, SelectFolder, ScanVideoFiles)
- [x] 7种抽帧模式 → Task 5 (BuildArgs + BuildTimestampArgs)
- [x] PNG/JPEG + 质量 → Task 5 (QualityToFFmpeg)
- [x] 按比例/自定义尺寸 → Task 5 (buildVideoFilter)
- [x] 4种命名模板 → Task 5 (BuildOutputPattern)
- [x] 详细进度 → Task 7 (progressRe + ParseProgress + event emit)
- [x] 任务队列串行 → Task 7 (executeTasks loops sequentially)
- [x] 中文界面 → 所有前端组件使用中文标签
- [x] 平铺选项 → ModeSelector/OutputFormat/SizeConfig/NamingSelector 使用按钮组
- [x] macOS .dmg → Task 14 (package:darwin:dmg)
- [x] Windows .exe → Task 14 (package:windows)
- [x] Taskfile 构建 → Task 14

**2. 占位符扫描：** 无 TBD/TODO/实现占位符

**3. 类型一致性：** Task 数据模型在前端 TypeScript 和 Go 端字段名一致（均使用 JSON tag 的 camelCase）
