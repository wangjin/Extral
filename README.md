# Extral

跨平台视频抽帧桌面客户端，基于 Wails 3 + React + Go 构建，内嵌 FFmpeg，双击即用。

## 功能

- **7种抽帧模式**：每N帧、每N秒、时间范围、总帧数、关键帧、场景切换、指定时间点
- **多格式输出**：JPEG（可调质量）、PNG
- **尺寸控制**：按比例缩放、自定义宽高（支持保持宽高比）
- **批量处理**：支持单文件、多文件、文件夹选择
- **任务队列**：串行执行，详细进度显示（进度条、帧数、剩余时间）
- **文件命名**：序号、时间戳、视频名+序号、视频名+时间

## 下载

前往 [Releases](../../releases) 下载最新版本。

| 平台 | 文件 |
|------|------|
| macOS | `Extral-macos.dmg` |
| Windows | `Extral-windows-amd64.exe` |

### macOS 使用说明

双击打开 .dmg 文件，将 Extral 拖入 Applications 文件夹即可。

如果提示"无法打开"或"已损坏"，请在终端执行：

```bash
xattr -cr /Applications/Extral.app
```

## 开发环境

- Go 1.26+
- Node.js 22+
- Wails 3 CLI：`go install github.com/wailsapp/wails/v3/cmd/wails3@latest`
- Taskfile：`go install github.com/go-task/task/v3/cmd/task@latest`
- go-winres：`go install github.com/tc-hib/go-winres@latest`（Windows 构建需要）

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
task test

# 生成 Wails 绑定
wails3 generate bindings -ts -d frontend/bindings .
```

## 构建与打包

```bash
# 当前平台构建
task build

# macOS .app + .dmg
task package:darwin:dmg

# Windows .exe（含图标）
task package:windows

# 清理构建产物
task clean
```

应用图标源文件位于 `assets/icon.png`，构建时自动生成 macOS `.icns` 和 Windows `.syso` 资源。

## CI/CD

- **PR 检查**：`.github/workflows/build.yml` — PR 到 main 时自动运行测试
- **发布构建**：`.github/workflows/release.yml` — 推送 `v*` 标签时自动构建 macOS/Windows 产物并创建 GitHub Release

```bash
# 发布新版本
git tag v1.0.0
git push origin v1.0.0
```

## 技术栈

- **后端**：Go + Wails 3
- **前端**：React 18 + TypeScript + Vite
- **视频处理**：FFmpeg（静态编译二进制内嵌）
