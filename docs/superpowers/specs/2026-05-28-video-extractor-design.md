# 视频抽帧桌面应用设计文档

## 概述

一个跨平台视频抽帧桌面客户端，基于 Wails 3 + React + Go 构建，内嵌 FFmpeg 二进制，双击即可使用，无需安装额外依赖。支持 Windows（exe）和 macOS（dmg）。

## 技术栈

- **后端**：Go + Wails 3 (`github.com/wailsapp/wails/v3`)
- **前端**：React + TypeScript + Vite
- **FFmpeg 集成**：内嵌 FFmpeg 静态编译二进制，通过 `os/exec` 调用
- **构建工具**：Taskfile（参考 Nearfy 项目结构）

## 功能需求

### 视频源选择

- 单文件选择
- 多文件批量选择
- 文件夹选择（自动识别视频文件）

### 抽帧模式（7种）

| 模式 | 说明 | 参数 |
|------|------|------|
| 每N帧 | 每隔 N 帧抽取一帧 | 间隔帧数 N |
| 每N秒 | 每隔 N 秒抽取一帧 | 间隔秒数 N |
| 时间范围 | 指定起止时间区间抽帧 | 开始时间、结束时间 |
| 总帧数 | 等间隔抽取指定数量的帧 | 目标帧数 |
| 关键帧 | 仅抽取视频编码关键帧(I帧) | 无 |
| 场景切换 | 自动在场景变化处抽帧 | 灵敏度阈值(默认0.3) |
| 指定时间点 | 在用户指定的精确时间点抽帧 | 时间点列表 |

### 输出格式

- JPEG（可调压缩质量 1-100，滑块控制）
- PNG（无损）

### 输出尺寸

- 按比例缩放（滑块，10%-200%）
- 自定义宽高（可勾选保持宽高比）

### 文件命名模板（平铺选项）

- 序号命名：`frame_0001.png`
- 时间戳命名：`00h01m30s_000f.png`
- 视频名+序号：`video1_0001.png`
- 视频名+时间戳：`video1_00h01m30s.png`

### 进度显示

详细进度信息：
- 进度条（百分比）
- 当前帧 / 预计总帧数
- 预计剩余时间
- 当前处理文件名（多文件时）

### 界面语言

纯中文界面。

## UI 设计

### 布局方案：任务队列模式

```
┌─────────────────────────────────────────┐
│ 顶部工具栏：标题 + [选择视频] [选择文件夹]  │
├─────────────────────────────────────────┤
│ 任务列表                                 │
│  ┌─ video1.mp4 ──── [进度条 65%] ──────┐ │
│  ├─ video2.mp4 ──── 等待中 ────────────┤ │
│  └─ video3.mp4 ──── 等待中 ────────────┘ │
├─────────────────────────────────────────┤
│ 配置面板（选中任务时显示）                  │
│  抽帧模式：[每N帧][每N秒][时间范围]...     │
│  输出格式：[JPEG] [PNG]  质量: ──●──      │
│  输出尺寸：[按比例] [自定义]  50%: ──●──   │
│  保存路径：[____________] [浏览]          │
│  文件命名：[序号][时间戳][视频名+序号]...  │
├─────────────────────────────────────────┤
│ 底部操作栏：统计信息 [全部开始] [清空队列]  │
└─────────────────────────────────────────┘
```

### UI 规则

- 所有选项平铺展示为按钮组，不使用下拉选择框
- 选中状态高亮显示
- 数值类参数使用滑块或输入框
- 配置面板为当前选中任务的配置，支持修改

## 架构设计

### 整体架构

```
┌─────────────────────────────────────────┐
│            Wails 3 应用                  │
│  ┌───────────────┐  ┌────────────────┐  │
│  │  React 前端    │  │   Go 后端      │  │
│  │  (UI渲染)      │←→│  (业务逻辑)    │  │
│  └───────────────┘  └───────┬────────┘  │
│                             │            │
│                     ┌───────┴───────┐    │
│                     │ FFmpeg 二进制  │    │
│                     │ (打包内嵌)     │    │
│                     └───────────────┘    │
└─────────────────────────────────────────┘
```

### 前后端通信

- 通过 Wails 3 bindings 机制，Go 方法直接暴露给前端
- 进度事件通过 `application.Get().Event.Emit()` 从后端推送到前端
- 前端使用 `@wailsio/runtime` 监听事件

### Go 服务层

- `App`（主服务）：Wails Service 生命周期管理，协调各子服务
- `internal/ffmpeg/`：FFmpeg 二进制路径管理、命令执行
- `internal/video/`：视频信息解析（ffprobe 获取时长、帧率、分辨率）
- `internal/extractor/`：抽帧逻辑（构建 FFmpeg 滤镜链、参数映射）
- `internal/task/`：任务队列管理（创建、排队、状态更新、串行执行）

### 数据模型

```go
type Task struct {
    ID        string
    Status    string    // pending / running / completed / failed / cancelled
    VideoPath string

    // 抽帧配置
    Mode   string      // every-frame / every-second / time-range / total-frames / keyframe / scene-change / timestamps
    Params ModeParams

    // 输出配置
    Format    string    // jpeg / png
    Quality   int       // 1-100
    ScaleMode string    // percentage / custom
    Scale     int       // 百分比（ScaleMode=percentage 时有效）
    Width     int       // 自定义宽度（ScaleMode=custom 时有效）
    Height    int       // 自定义高度（ScaleMode=custom 时有效）
    KeepRatio bool      // 自定义模式下是否保持宽高比

    // 输出路径
    OutputDir string
    Naming    string    // sequential / timestamp / video-seq / video-time

    // 运行时状态
    Progress    float64
    Current     int       // 已抽帧数
    Total       int       // 预计总帧数
    ElapsedTime int64     // 秒
    Error       string
    CreatedAt   time.Time
}

type ModeParams struct {
    // every-frame
    FrameInterval int
    // every-second
    SecondInterval float64
    // time-range
    StartTime string
    EndTime   string
    // total-frames
    TotalFrames int
    // scene-change
    SceneThreshold float64 // 默认 0.3
    // timestamps
    Timestamps []string
}
```

### 任务流程

```
用户添加视频 → 创建 Task（默认配置）→ 加入队列
      ↓
用户选中任务 → 配置面板修改参数 → 更新 Task
      ↓
点击"全部开始" → 按 FIFO 串行执行
      ↓
执行：ffprobe 解析视频信息 → 计算 Total → 启动 ffmpeg
      ↓
解析 ffmpeg stderr → 推送进度事件 → 前端更新
      ↓
完成 → 标记 completed → 自动执行下一个 pending 任务
```

任务串行执行，避免多个 FFmpeg 进程同时占用系统资源。

## FFmpeg 命令映射

| 模式 | FFmpeg 命令 |
|------|------------|
| 每N帧 | `ffmpeg -i input -vf "select='not(mod(n\,N))'" -vsync vfr output_%04d.jpg` |
| 每N秒 | `ffmpeg -i input -vf "fps=1/N" output_%04d.jpg` |
| 时间范围 | `ffmpeg -ss START -to END -i input output_%04d.jpg`（抽取范围内全部帧） |
| 总帧数 | `ffmpeg -i input -vf "select='not(mod(n\,round(total_frames/N)))'" -vsync vfr -frames:v N output_%04d.jpg` |
| 关键帧 | `ffmpeg -i input -vf "select='eq(pict_type\,I)'" -vsync vfr output_%04d.jpg` |
| 场景切换 | `ffmpeg -i input -vf "select='gt(scene\,0.3)'" -vsync vfr output_%04d.jpg` |
| 指定时间点 | 对每个时间点：`ffmpeg -ss T -i input -frames:v 1 output.jpg` |

### 滤镜链扩展

- **尺寸缩放**：追加 `scale=iw*0.5:ih*0.5`（按比例）或 `scale=1920:1080`（自定义），保持宽高比用 `scale=W:-1`
- **JPEG 质量**：`-q:v` 参数（1最好31最差，映射百分比 1-100）

### 进度解析

通过解析 ffmpeg stderr 中的 `frame=` 和 `time=` 行，结合总时长计算进度百分比。

## 项目结构

```
video-extractor/
├── build/                     # 构建资源
│   ├── darwin/
│   │   ├── icons.icns
│   │   └── Info.plist
│   └── appicon.png
├── thirdparty/                # FFmpeg 静态编译二进制
│   ├── darwin-arm64/
│   │   ├── ffmpeg
│   │   └── ffprobe
│   ├── darwin-amd64/
│   │   ├── ffmpeg
│   │   └── ffprobe
│   └── windows-amd64/
│       ├── ffmpeg.exe
│       └── ffprobe.exe
├── frontend/                  # React 前端
│   ├── src/
│   │   ├── App.tsx
│   │   ├── App.css
│   │   ├── main.tsx
│   │   ├── components/
│   │   │   ├── TaskList.tsx
│   │   │   ├── TaskItem.tsx
│   │   │   ├── ConfigPanel.tsx
│   │   │   ├── ModeSelector.tsx
│   │   │   ├── OutputFormat.tsx
│   │   │   ├── SizeConfig.tsx
│   │   │   ├── NamingSelector.tsx
│   │   │   └── ActionBar.tsx
│   │   ├── hooks/
│   │   │   └── useTaskQueue.ts
│   │   └── bindings/
│   ├── index.html
│   ├── package.json
│   ├── vite.config.ts
│   └── tsconfig.json
├── internal/
│   ├── ffmpeg/
│   │   └── binary.go          # FFmpeg 路径管理、命令执行
│   ├── video/
│   │   └── info.go            # ffprobe 视频信息解析
│   ├── extractor/
│   │   └── extractor.go       # 抽帧逻辑、滤镜链构建
│   └── task/
│       └── manager.go         # 任务队列管理
├── main.go
├── app.go
├── Taskfile.yml
├── wails.json
├── go.mod
└── .gitignore
```

## 打包方案

### FFmpeg 分发策略

不使用 `go:embed`（二进制过大），改为构建时复制：

- **macOS**：FFmpeg 二进制放入 `.app/Contents/Resources/ffmpeg/`
- **Windows**：FFmpeg 放在 `.exe` 同级目录

运行时通过可执行文件路径推算 FFmpeg 位置。

### 构建（参考 Nearfy 的 Taskfile 模式）

- `task build` — 当前平台构建
- `task build:darwin:arm64` — macOS Apple Silicon
- `task build:windows:amd64` — Windows
- `task package:darwin:dmg` — 打包 macOS .app + .dmg
- `task package:windows` — 打包 Windows .exe

### macOS 打包流程

1. `go build` 生成二进制
2. 组装 `.app` bundle（Info.plist + icons + 二进制 + FFmpeg）
3. `codesign --force --deep --sign -` 签名
4. `hdiutil` 生成 `.dmg`

### Windows 打包流程

1. `go build -ldflags "-H windowsgui"` 生成无控制台窗口的 .exe
2. 复制 FFmpeg 到同目录
3. 分发为 zip 或 NSIS 安装包

## 不在范围内

- 视频播放/预览功能
- 抽帧结果预览
- 配置模板保存/加载
- 自动更新
- 多语言支持（仅中文）
