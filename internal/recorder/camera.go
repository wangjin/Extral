package recorder

import (
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"video-extractor/internal/ffmpeg"
)

var (
	avfDeviceRe    = regexp.MustCompile(`\[(\d+)\]\s+(.+)`)
	profilerNameRe = regexp.MustCompile(`^\s{8,12}(\S[^:]+):$`)
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
		if strings.Contains(trimmed, `"DirectShow video devices"`) {
			section = "video"
			continue
		}
		if strings.Contains(trimmed, `"DirectShow audio devices"`) {
			section = "audio"
			continue
		}
		// Skip "Alternative name" lines
		if strings.Contains(trimmed, "Alternative name") {
			continue
		}
		// Extract device name from the last [brackets] on the line
		lastOpen := strings.LastIndex(trimmed, "[")
		if lastOpen < 0 {
			continue
		}
		closeIdx := strings.Index(trimmed[lastOpen:], "]")
		if closeIdx < 0 {
			continue
		}
		name := trimmed[lastOpen+1 : lastOpen+closeIdx]
		// Skip the dshow prefix line (contains @)
		if strings.Contains(name, "@") {
			continue
		}
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
