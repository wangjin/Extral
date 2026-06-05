package recorder

import (
	"runtime"
	"strings"
	"testing"
)

func TestBuildPreviewArgs(t *testing.T) {
	camera := Camera{ID: "0", Name: "FaceTime HD Camera"}
	args := BuildPreviewArgs(camera)

	if runtime.GOOS == "darwin" {
		found := false
		for i, a := range args {
			if a == "-f" && i+1 < len(args) && args[i+1] == "avfoundation" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected -f avfoundation for darwin")
		}
	}

	// Common args for all platforms
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "image2pipe") {
		t.Error("expected image2pipe muxer")
	}
	if !strings.Contains(joined, "pipe:1") {
		t.Error("expected pipe:1 output")
	}
	if !strings.Contains(joined, "scale=640:-1") {
		t.Error("expected scale=640:-1 filter")
	}
}

func TestBuildRecordingArgs_CRF(t *testing.T) {
	cfg := RecordingConfig{
		CameraID:    "0",
		AudioDevice: "",
		Resolution:  "1280x720",
		FPS:         30,
		Codec:       "libx264",
		RateControl: "crf",
		Quality:     "medium",
		CRF:         23,
		Format:      "mp4",
		MaxDuration: 0,
	}
	camera := Camera{ID: "0", Name: "FaceTime HD Camera"}
	args := BuildRecordingArgs(cfg, camera, "/tmp/out.mp4")

	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "-c:v libx264") {
		t.Errorf("expected -c:v libx264 in args, got %s", joined)
	}
	if !strings.Contains(joined, "-crf 23") {
		t.Errorf("expected -crf 23 in args, got %s", joined)
	}
	if strings.Contains(joined, "-b:v") {
		t.Error("did not expect -b:v in CRF mode")
	}
	if !strings.Contains(joined, "/tmp/out.mp4") {
		t.Error("expected output path in args")
	}
	if !strings.Contains(joined, "-preset medium") {
		t.Error("expected -preset medium")
	}
}

func TestBuildRecordingArgs_Bitrate(t *testing.T) {
	cfg := RecordingConfig{
		CameraID:    "0",
		AudioDevice: "1",
		Resolution:  "1920x1080",
		FPS:         60,
		Codec:       "libx264",
		RateControl: "bitrate",
		Bitrate:     8000,
		Format:      "mp4",
	}
	camera := Camera{ID: "0", Name: "FaceTime HD Camera"}
	args := BuildRecordingArgs(cfg, camera, "/tmp/out.mp4")

	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "-b:v 8000k") {
		t.Errorf("expected -b:v 8000k in args, got %s", joined)
	}
	if strings.Contains(joined, "-crf") {
		t.Error("did not expect -crf in bitrate mode")
	}
	if !strings.Contains(joined, "-c:a aac") {
		t.Error("expected -c:a aac when audio device is set")
	}
	if !strings.Contains(joined, "scale=1920:1080") {
		t.Error("expected scale filter for 1080p")
	}
}

func TestBuildRecordingArgs_MaxDuration(t *testing.T) {
	cfg := RecordingConfig{
		CameraID:    "0",
		Resolution:  "1280x720",
		FPS:         30,
		Codec:       "libx264",
		RateControl: "crf",
		CRF:         23,
		Format:      "mp4",
		MaxDuration: 120,
	}
	camera := Camera{ID: "0", Name: "FaceTime HD Camera"}
	args := BuildRecordingArgs(cfg, camera, "/tmp/out.mp4")

	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "-t 120") {
		t.Errorf("expected -t 120 in args, got %s", joined)
	}
}

func TestBuildRecordingArgs_NoAudio(t *testing.T) {
	cfg := RecordingConfig{
		CameraID:    "0",
		AudioDevice: "",
		Resolution:  "1280x720",
		FPS:         30,
		Codec:       "libx264",
		RateControl: "crf",
		CRF:         23,
		Format:      "mp4",
	}
	camera := Camera{ID: "0", Name: "FaceTime HD Camera"}
	args := BuildRecordingArgs(cfg, camera, "/tmp/out.mp4")

	joined := strings.Join(args, " ")
	if strings.Contains(joined, "-c:a") {
		t.Error("did not expect -c:a when no audio device")
	}
}

func TestQualityToCRF(t *testing.T) {
	tests := []struct {
		quality string
		want    int
	}{
		{"low", 28},
		{"medium", 23},
		{"high", 18},
		{"unknown", 23},
	}
	for _, tt := range tests {
		got := QualityToCRF(tt.quality)
		if got != tt.want {
			t.Errorf("QualityToCRF(%q) = %d, want %d", tt.quality, got, tt.want)
		}
	}
}
