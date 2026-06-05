package recorder

import (
	"context"
	"encoding/base64"
	"io"
	"os/exec"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"

	"video-extractor/internal/ffmpeg"
)

// Preview manages a low-fps MJPEG preview stream from a camera.
type Preview struct {
	mu     sync.Mutex
	ffmpeg *ffmpeg.Manager
	cmd    *exec.Cmd
	cancel context.CancelFunc
	camera Camera
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
			application.Get().Event.Emit("recorder:log", string(buf[:n]))
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
