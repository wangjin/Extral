# 视频抽帧工具

基于 Wails 3 + React + Go 构建的跨平台视频抽帧桌面客户端，内嵌 FFmpeg，双击即用。

## 功能

- **7种抽帧模式**：每N帧、每N秒、时间范围、总帧数、关键帧、场景切换、指定时间点
- **多格式输出**：JPEG（可调质量）、PNG
- **尺寸控制**：按比例缩放、自定义宽高（支持保持宽高比）
- **批量处理**：支持单文件、多文件、文件夹选择
- **任务队列**：串行执行，详细进度显示（进度条、帧数、剩余时间）
- **文件命名**：序号、时间戳、视频名+序号、视频名+时间

## 开发环境

- Go 1.25+
- Node.js 16+
- Wails 3 CLI：`go install github.com/wailsapp/wails/v3/cmd/wails3@latest`
- Taskfile：`go install github.com/go-task/task/v3/cmd/task@latest`

## 准备 FFmpeg 二进制

从以下地址下载对应平台的静态编译版本，解压后放入 `thirdparty/` 目录：

### macOS ARM64 (Apple Silicon)

```
https://evermeet.cx/ffmpeg/getrelease/ffmpeg/arm64/zip
https://evermeet.cx/ffmpeg/getrelease/ffprobe/arm64/zip
```

放入：`thirdparty/darwin-arm64/`

### macOS AMD64 (Intel)

```
https://evermeet.cx/ffmpeg/getrelease/ffmpeg/x86_64/zip
https://evermeet.cx/ffmpeg/getrelease/ffprobe/x86_64/zip
```

放入：`thirdparty/darwin-amd64/`

### Windows AMD64

```
https://www.gyan.dev/ffmpeg/builds/
```

下载 `ffmpeg-release-essentials.zip`，解压 `bin/` 目录下的 `ffmpeg.exe` 和 `ffprobe.exe`。

放入：`thirdparty/windows-amd64/`

### 目录结构

```
thirdparty/
├── darwin-arm64/
│   ├── ffmpeg
│   └── ffprobe
├── darwin-amd64/
│   ├── ffmpeg
│   └── ffprobe
└── windows-amd64/
    ├── ffmpeg.exe
    └── ffprobe.exe
```

## 开发

```bash
# 启动开发模式（热重载）
wails3 dev

# 运行测试
go test ./internal/... -v

# 生成 Wails 绑定
wails3 generate bindings -ts -d frontend/bindings .
```

## 构建与打包

```bash
# 当前平台构建
task build

# macOS .app + .dmg
task package:darwin:dmg

# Windows .exe
task package:windows

# 清理构建产物
task clean
```

## 技术栈

- **后端**：Go + Wails 3
- **前端**：React 18 + TypeScript + Vite
- **视频处理**：FFmpeg（静态编译二进制内嵌）
