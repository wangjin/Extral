package app

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"

	"video-extractor/internal/extractor"
	"video-extractor/internal/ffmpeg"
	"video-extractor/internal/model"
	"video-extractor/internal/task"
	"video-extractor/internal/video"
)

type Service struct {
	ctx       context.Context
	taskMgr   *task.Manager
	ffmpegMgr *ffmpeg.Manager
	version   string
}

func NewService(version string) *Service {
	return &Service{
		taskMgr:   task.NewManager(),
		ffmpegMgr: ffmpeg.NewManager(),
		version:   version,
	}
}

func (s *Service) GetVersion() string {
	return s.version
}

func (s *Service) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	s.ctx = ctx
	ffmpeg.ExtractEmbedded()
	s.taskMgr.SetCallback(func(t *model.Task) {
		application.Get().Event.Emit("task:changed", t)
	})
	return nil
}

func (s *Service) ServiceShutdown() error {
	return nil
}

func (s *Service) AddTasks(paths []string) []*model.Task {
	var tasks []*model.Task
	for _, p := range paths {
		t := s.taskMgr.AddTask(p)
		videoName := strings.TrimSuffix(filepath.Base(p), filepath.Ext(p))
		t.OutputDir = filepath.Join(filepath.Dir(p), videoName+"_frames")
		s.taskMgr.UpdateTask(t)
		tasks = append(tasks, t)
	}
	return tasks
}

func (s *Service) GetTasks() []*model.Task {
	return s.taskMgr.ListTasks()
}

func (s *Service) UpdateTask(t *model.Task) {
	s.taskMgr.UpdateTask(t)
}

func (s *Service) RemoveTask(id string) {
	s.taskMgr.RemoveTask(id)
}

func (s *Service) ClearTasks() {
	s.taskMgr.ClearTasks()
}

func (s *Service) SelectFiles() []string {
	result, err := application.Get().Dialog.OpenFile().
		SetTitle("选择视频文件").
		PromptForMultipleSelection()
	if err != nil || len(result) == 0 {
		return nil
	}
	return result
}

func (s *Service) SelectFolder() string {
	result, err := application.Get().Dialog.OpenFile().
		SetTitle("选择包含视频的文件夹").
		PromptForSingleSelection()
	if err != nil || result == "" {
		return ""
	}
	return result
}

func (s *Service) SelectOutputDir() string {
	result, err := application.Get().Dialog.OpenFile().
		SetTitle("选择输出保存路径").
		PromptForSingleSelection()
	if err != nil || result == "" {
		return ""
	}
	return result
}

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

func (s *Service) StartAll() {
	pending := s.taskMgr.GetPendingTasks()
	go s.executeTasks(pending)
}

func (s *Service) CancelTask(id string) {
	t := s.taskMgr.GetTask(id)
	if t != nil && t.Status == model.StatusRunning {
		t.Status = model.StatusCancelled
		s.taskMgr.UpdateTask(t)
	}
}

func (s *Service) executeTasks(tasks []*model.Task) {
	for _, t := range tasks {
		if t.Status == model.StatusCancelled {
			continue
		}
		s.executeTask(t)
	}
}

func (s *Service) executeTask(t *model.Task) {
	info, err := s.GetVideoInfo(t.VideoPath)
	if err != nil {
		t.Status = model.StatusFailed
		t.Error = fmt.Sprintf("获取视频信息失败: %v", err)
		s.taskMgr.UpdateTask(t)
		return
	}

	if err := os.MkdirAll(t.OutputDir, 0755); err != nil {
		t.Status = model.StatusFailed
		t.Error = fmt.Sprintf("创建输出目录失败: %v", err)
		s.taskMgr.UpdateTask(t)
		return
	}

	t.Status = model.StatusRunning
	t.Total = info.TotalFrames
	t.Logs = []string{}
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

	ffmpegPath, _ := s.ffmpegMgr.Paths()
	s.taskMgr.AppendLog(t.ID, fmt.Sprintf("[命令] %s %s", ffmpegPath, strings.Join(args, " ")))

	cmd := s.ffmpegMgr.RunFFmpeg(args...)
	stderr, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		t.Status = model.StatusFailed
		t.Error = fmt.Sprintf("启动FFmpeg失败: %v", err)
		s.taskMgr.AppendLog(t.ID, fmt.Sprintf("[错误] %v", err))
		s.taskMgr.UpdateTask(t)
		return
	}

	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		line := scanner.Text()
		s.taskMgr.AppendLog(t.ID, line)

		frame, _, ok := ffmpeg.ParseProgress(line)
		if ok && t.Total > 0 {
			t.Progress = float64(frame) / float64(t.Total) * 100
			t.Current = frame
			t.ElapsedTime = int64(time.Since(startTime).Seconds())
			s.taskMgr.UpdateTask(t)
		}

		current := s.taskMgr.GetTask(t.ID)
		if current != nil && current.Status == model.StatusCancelled {
			cmd.Process.Kill()
			return
		}
	}
	waitErr := cmd.Wait()

	current := s.taskMgr.GetTask(t.ID)
	if current != nil && current.Status == model.StatusCancelled {
		return
	}

	if waitErr != nil {
		t.Status = model.StatusFailed
		t.Error = fmt.Sprintf("FFmpeg 退出异常: %v", waitErr)
		s.taskMgr.AppendLog(t.ID, fmt.Sprintf("[退出码] %v", waitErr))
		s.taskMgr.UpdateTask(t)
		return
	}

	t.Status = model.StatusCompleted
	t.Progress = 100
	t.ElapsedTime = int64(time.Since(startTime).Seconds())
	s.taskMgr.AppendLog(t.ID, "[完成]")
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

		ffmpegPath, _ := s.ffmpegMgr.Paths()
		s.taskMgr.AppendLog(t.ID, fmt.Sprintf("[命令 %d/%d] %s %s", i+1, totalCmds, ffmpegPath, strings.Join(args, " ")))

		cmd := s.ffmpegMgr.RunFFmpeg(args...)
		out, runErr := cmd.CombinedOutput()
		if runErr != nil {
			s.taskMgr.AppendLog(t.ID, string(out))
			s.taskMgr.AppendLog(t.ID, fmt.Sprintf("[错误] %v", runErr))
		}

		t.Progress = float64(i+1) / float64(totalCmds) * 100
		t.Current = i + 1
		t.ElapsedTime = int64(time.Since(startTime).Seconds())
		s.taskMgr.UpdateTask(t)
	}

	current := s.taskMgr.GetTask(t.ID)
	if current != nil && current.Status == model.StatusCancelled {
		return
	}

	t.Status = model.StatusCompleted
	t.Progress = 100
	t.ElapsedTime = int64(time.Since(startTime).Seconds())
	s.taskMgr.AppendLog(t.ID, "[完成]")
	s.taskMgr.UpdateTask(t)
}
