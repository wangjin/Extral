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
