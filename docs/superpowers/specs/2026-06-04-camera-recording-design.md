# 摄像头录制功能设计

## 概述

为 Extral（视频抽帧桌面客户端）新增独立的摄像头视频录制功能模块。用户可以通过 Tab 切换在「抽帧」和「录制」两个功能之间切换。录制功能通过 FFmpeg 直接采集摄像头画面并编码保存，支持实时预览、音频录制、多种参数配置。

## 1. 整体架构

录制功能作为独立模块，与现有抽帧功能完全解耦：

- **Go 后端**：新增 `internal/recorder` 包，注册为独立 Wails Service
- **前端**：工具栏 Tab 切换，录制界面为左右分栏布局
- **FFmpeg**：重新编译，新增摄像头采集设备和 H.264 编码能力

### 1.1 新增文件

```
internal/recorder/
├── service.go     # RecorderService — Wails 服务注册入口
├── camera.go      # 摄像头/音频设备枚举（FFprobe + 原生兜底）
├── session.go     # 录制会话管理（开始/暂停/停止/状态）
├── preview.go     # FFmpeg 预览帧流 → Wails 事件推送
└── models.go      # 数据模型（Camera、RecordingConfig、RecordingState）

frontend/src/
├── components/
│   └── RecorderPanel.tsx    # 录制主面板（左右分栏布局）
├── hooks/
│   └── useRecorder.ts       # 录制状态管理 hook
```

### 1.2 main.go 集成

```go
recorderService := recorder.NewService()
application.NewService(recorderService),
```

`recorder.Service` 独立注册，有自己的 `ServiceStartup` / `ServiceShutdown` 生命周期。`ServiceShutdown` 时自动停止所有进行中的录制和预览进程。

## 2. 后端数据模型

### 2.1 设备模型

```go
type Camera struct {
    ID   string `json:"id"`   // FFmpeg 设备标识，如 "0"
    Name string `json:"name"` // 显示名称，如 "FaceTime HD Camera"
}

type AudioDevice struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}
```

### 2.2 录制配置

```go
type RecordingConfig struct {
    CameraID    string `json:"cameraId"`
    AudioDevice string `json:"audioDevice"`  // 空字符串=不录音频, "none"=明确无音频
    Resolution  string `json:"resolution"`   // "1920x1080" | "1280x720" | "854x480"
    FPS         int    `json:"fps"`          // 30 | 60
    Codec       string `json:"codec"`        // "libx264"
    RateControl string `json:"rateControl"`  // "crf" | "bitrate"
    Quality     string `json:"quality"`      // "low" | "medium" | "high" | "custom"
    CRF         int    `json:"crf"`          // CRF 模式下使用, 0-51
    Bitrate     int    `json:"bitrate"`      // 码率模式下使用, kbps
    Format      string `json:"format"`       // "mp4" | "mkv"
    MaxDuration int    `json:"maxDuration"`  // 秒, 0=不限
    OutputDir   string `json:"outputDir"`
    FileName    string `json:"fileName"`     // 自动生成，可修改
}
```

### 2.3 录制状态

```go
type RecordingStatus struct {
    State     string  `json:"state"`     // "idle" | "previewing" | "recording" | "paused"
    Duration  int64   `json:"duration"`  // 已录秒数
    FileSize  int64   `json:"fileSize"`  // 字节
    FPS       float64 `json:"fps"`       // 实际帧率
    Dropped   int     `json:"dropped"`   // 丢帧数
    Error     string  `json:"error"`     // 错误信息
}
```

## 3. 后端 Service API

```go
// 设备枚举
func (s *Service) ListCameras() ([]Camera, error)
func (s *Service) ListAudioDevices() ([]AudioDevice, error)

// 预览控制
func (s *Service) StartPreview(cameraID string) error
func (s *Service) StopPreview() error

// 录制控制
func (s *Service) StartRecording(config RecordingConfig) error
func (s *Service) StopRecording() error
func (s *Service) PauseRecording() error   // 仅 macOS/Linux
func (s *Service) ResumeRecording() error  // 仅 macOS/Linux

// 文件选择
func (s *Service) SelectOutputDir() string
```

### 3.1 Wails 事件

| 事件名 | 数据 | 用途 |
|--------|------|------|
| `recorder:preview-frame` | base64 JPEG 字符串 | 预览帧更新 |
| `recorder:status` | RecordingStatus 对象 | 录制状态更新 |
| `recorder:log` | string | FFmpeg 实时日志行 |
| `recorder:error` | `{error: string}` | 错误通知 |

## 4. 设备枚举

### 4.1 视频设备

**macOS**（优先 FFprobe）：
```bash
ffprobe -list_devices true -f avfoundation -i dummy 2>&1
```
解析 `[AVFoundation input device]` 段落中 `[video]` 标记的设备。

**macOS 兜底**（FFprobe 解析失败时）：
```bash
system_profiler SPCameraDataType 2>/dev/null
```
解析设备名，ID 按顺序 0, 1, 2...

**Windows**：
```bash
ffprobe -list_devices true -f dshow -i dummy 2>&1
```
解析 `[dshow]` 段落中 `[video]` 标记的设备。dshow 设备信息通常准确，无需兜底。

### 4.2 音频设备

同理，解析 `[audio]` 段的设备。

## 5. FFmpeg 命令构建

### 5.1 预览帧流

```bash
# macOS
ffmpeg -f avfoundation -framerate 5 -i "<cameraID>" \
  -vf "scale=640:-1" -f image2pipe -vcodec mjpeg pipe:1

# Windows
ffmpeg -f dshow -i "video=<cameraName>" \
  -vf "scale=640:-1" -f image2pipe -vcodec mjpeg pipe:1
```

以 5fps 低帧率截取，缩放到 640px 宽节省带宽。Go 端从 stdout 循环读取 JPEG 帧（按 FFmpeg DCLI 边界分割），每收到一帧通过 Wails 事件推送。

### 5.2 录制命令

**CRF 模式（默认）**：
```bash
# macOS（含音频）
ffmpeg -f avfoundation -framerate <fps> -i "<videoID>:<audioID>" \
  -c:v libx264 -crf <crf> -preset medium \
  -c:a aac -b:a 128k \
  -t <maxDuration> \
  -y <outputPath>

# macOS（仅视频）
ffmpeg -f avfoundation -framerate <fps> -i "<videoID>:none" \
  -c:v libx264 -crf <crf> -preset medium \
  -t <maxDuration> \
  -y <outputPath>

# Windows（含音频）
ffmpeg -f dshow -framerate <fps> -i "video=<cameraName>:audio=<audioName>" \
  -c:v libx264 -crf <crf> -preset medium \
  -c:a aac -b:a 128k \
  -t <maxDuration> \
  -y <outputPath>
```

**码率模式**：`-crf <crf>` 替换为 `-b:v <bitrate>k`。

### 5.3 CRF 质量映射

| Quality | CRF 值 |
|---------|--------|
| low     | 28     |
| medium  | 23     |
| high    | 18     |
| custom  | 用户自定义 (0-51) |

### 5.4 分辨率映射

| 选项      | 实际值       |
|-----------|-------------|
| 1080p     | 1920x1080   |
| 720p      | 1280x720    |
| 480p      | 854x480     |

通过 `-vf scale=W:H` 实现。

### 5.5 录制状态解析

从 FFmpeg stderr 解析进度行，复用现有 `ffmpeg.ParseProgress`，额外解析：
- `size=` → 文件大小
- `fps=` → 实际帧率
- `drop=` → 丢帧数

## 6. 暂停/恢复

**macOS / Linux**：向 FFmpeg 进程发送 `SIGTSTP`（暂停）和 `SIGCONT`（恢复），进程挂起但保持文件句柄。

**Windows**：不支持信号暂停。前端根据平台隐藏暂停按钮。

## 7. 前端设计

### 7.1 Tab 切换

工具栏中 logo 右侧新增两个 Tab：「抽帧」|「录制」。`App.tsx` 用 `activeTab: 'extract' | 'record'` 状态控制切换。

切换到「录制」时，TaskList + ConfigPanel 区域替换为 RecorderPanel。ActionBar 替换为录制专属的状态栏。

### 7.2 RecorderPanel 布局（左右分栏）

**左侧**：
- 摄像头预览区（`<img>` 绑定 base64 帧）
- 录制控制按钮：开始录制 / 暂停 / 停止
- 状态栏：录制时长、文件大小、实际帧率

**右侧**：参数配置面板
- 摄像头下拉选择
- 麦克风下拉选择
- 分辨率选择（1080p / 720p / 480p pill 按钮组）
- 帧率选择（30 / 60 pill 按钮组）
- 编码器（固定 H.264）
- 质量控制：CRF / 码率 模式切换
  - CRF 模式：低 / 中 / 高 / 自定义（滑块）
  - 码率模式：目标码率输入框（kbps）
- 输出格式（MP4 / MKV pill 按钮组）
- 最大录制时长（可选）
- 保存路径（自动生成文件名 + 浏览按钮修改）

### 7.3 useRecorder hook

```typescript
interface RecorderState {
  cameras: Camera[]
  audioDevices: AudioDevice[]
  selectedCamera: string
  selectedAudio: string
  config: RecordingConfig
  previewFrame: string        // base64 JPEG
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
```

### 7.4 生命周期

- 切换到录制 Tab → 自动调用 `ListCameras()` + `ListAudioDevices()`
- 选择摄像头 → 自动启动预览 `StartPreview()`
- 离开录制 Tab → 停止预览释放设备 `StopPreview()`
- 开始录制 → `StopPreview()` 暂停预览（释放设备），然后 `StartRecording()`
- 录制期间状态通过 Wails 事件实时更新
- 停止录制 → `StopRecording()`，可选重新启动预览

## 8. 错误处理

| 场景 | 处理方式 |
|------|---------|
| 无摄像头 | 前端显示空状态「未检测到摄像头设备」 |
| 摄像头被占用 | `recorder:error` 事件，前端弹错误提示 |
| 录制中拔出摄像头 | FFmpeg 进程退出，Go 端检测退出码非零，标记失败 |
| 磁盘空间不足 | FFmpeg 写入失败退出，同上 |
| FFmpeg 不存在 | 复用 `FFmpegExists()` 检查，禁用录制 Tab |

## 9. FFmpeg 编译更新

现有 FFmpeg 使用 `--disable-everything` 精简编译，仅支持抽帧。录制功能需要新增：

```
--enable-avfoundation              # macOS 摄像头/音频采集
--enable-indev=avfoundation        # macOS 输入设备
--enable-dshow                     # Windows 摄像头/音频采集
--enable-indev=dshow               # Windows 输入设备
--enable-protocol=pipe             # 预览帧 pipe 输出
--enable-muxer=mp4                 # MP4 输出
--enable-muxer=matroska            # MKV 输出（已有 demuxer，补充 muxer）
--enable-muxer=image2pipe          # 预览帧 JPEG 流
--enable-encoder=libx264           # H.264 视频编码（需依赖 libx264）
--enable-encoder=aac               # AAC 音频编码
--enable-encoder=pcm_s16le         # 音频采集
--enable-decoder=rawvideo          # 原始视频解码（预览管道）
--enable-filter=scale              # 已有，保留
```

macOS 编译需预装 libx264（`brew install x264` 或从源码编译）。Windows 使用 MSYS2 环境编译 libx264 静态库。

预期二进制体积从 ~5MB 增长到 ~15-20MB。
