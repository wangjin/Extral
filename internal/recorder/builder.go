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
