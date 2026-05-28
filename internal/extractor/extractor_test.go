package extractor

import (
	"strings"
	"testing"

	"video-extractor/internal/model"
)

func TestBuildArgs_EveryFrame(t *testing.T) {
	task := &model.Task{
		Mode: model.ModeEveryFrame,
		Params: model.ModeParams{
			FrameInterval: 10,
		},
		Format:  "jpeg",
		Quality: 80,
	}
	args, err := BuildArgs(task, "/input/video.mp4", "/output")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	argStr := strings.Join(args, " ")
	if !strings.Contains(argStr, "select='not(mod(n\\,10))'") {
		t.Errorf("expected select filter for every 10 frames, got: %s", argStr)
	}
	if !strings.Contains(argStr, "-vsync") || !strings.Contains(argStr, "vfr") {
		t.Error("expected -vsync vfr for every-frame mode")
	}
}

func TestBuildArgs_EverySecond(t *testing.T) {
	task := &model.Task{
		Mode: model.ModeEverySecond,
		Params: model.ModeParams{
			SecondInterval: 2,
		},
		Format: "jpeg",
	}
	args, err := BuildArgs(task, "/input/video.mp4", "/output")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	argStr := strings.Join(args, " ")
	if !strings.Contains(argStr, "fps=1/2") {
		t.Errorf("expected fps=1/2, got: %s", argStr)
	}
}

func TestBuildArgs_TimeRange(t *testing.T) {
	task := &model.Task{
		Mode: model.ModeTimeRange,
		Params: model.ModeParams{
			StartTime: "00:00:10",
			EndTime:   "00:00:30",
		},
		Format: "jpeg",
	}
	args, err := BuildArgs(task, "/input/video.mp4", "/output")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	argStr := strings.Join(args, " ")
	if !strings.Contains(argStr, "-ss") || !strings.Contains(argStr, "00:00:10") {
		t.Errorf("expected -ss 00:00:10, got: %s", argStr)
	}
	if !strings.Contains(argStr, "-to") || !strings.Contains(argStr, "00:00:30") {
		t.Errorf("expected -to 00:00:30, got: %s", argStr)
	}
}

func TestBuildArgs_TotalFrames(t *testing.T) {
	task := &model.Task{
		Mode: model.ModeTotalFrames,
		Params: model.ModeParams{
			TotalFrames: 20,
		},
		Format: "jpeg",
	}
	args, err := BuildArgs(task, "/input/video.mp4", "/output")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	argStr := strings.Join(args, " ")
	if !strings.Contains(argStr, "-frames:v") || !strings.Contains(argStr, "20") {
		t.Errorf("expected -frames:v 20, got: %s", argStr)
	}
}

func TestBuildArgs_Keyframe(t *testing.T) {
	task := &model.Task{
		Mode:   model.ModeKeyframe,
		Format: "jpeg",
	}
	args, err := BuildArgs(task, "/input/video.mp4", "/output")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	argStr := strings.Join(args, " ")
	if !strings.Contains(argStr, "eq(pict_type\\,I)") {
		t.Errorf("expected keyframe filter, got: %s", argStr)
	}
}

func TestBuildArgs_SceneChange(t *testing.T) {
	task := &model.Task{
		Mode: model.ModeSceneChange,
		Params: model.ModeParams{
			SceneThreshold: 0.4,
		},
		Format: "jpeg",
	}
	args, err := BuildArgs(task, "/input/video.mp4", "/output")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	argStr := strings.Join(args, " ")
	if !strings.Contains(argStr, "gt(scene\\,0.4)") {
		t.Errorf("expected scene change filter with threshold 0.4, got: %s", argStr)
	}
}

func TestBuildArgs_ScalePercentage(t *testing.T) {
	task := &model.Task{
		Mode:      model.ModeEverySecond,
		Format:    "jpeg",
		ScaleMode: "percentage",
		Scale:     50,
	}
	args, err := BuildArgs(task, "/input/video.mp4", "/output")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	argStr := strings.Join(args, " ")
	if !strings.Contains(argStr, "scale=iw*0.50:ih*0.50") {
		t.Errorf("expected scale filter, got: %s", argStr)
	}
}

func TestBuildArgs_ScaleCustom(t *testing.T) {
	task := &model.Task{
		Mode:      model.ModeEverySecond,
		Format:    "jpeg",
		ScaleMode: "custom",
		Width:     1920,
		Height:    1080,
		KeepRatio: true,
	}
	args, err := BuildArgs(task, "/input/video.mp4", "/output")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	argStr := strings.Join(args, " ")
	if !strings.Contains(argStr, "scale=1920:-1") {
		t.Errorf("expected scale=1920:-1, got: %s", argStr)
	}
}

func TestBuildArgs_JPEGQuality(t *testing.T) {
	task := &model.Task{
		Mode:    model.ModeEverySecond,
		Format:  "jpeg",
		Quality: 80,
	}
	args, err := BuildArgs(task, "/input/video.mp4", "/output")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	argStr := strings.Join(args, " ")
	if !strings.Contains(argStr, "-q:v") {
		t.Errorf("expected -q:v for jpeg quality, got: %s", argStr)
	}
}

func TestBuildArgs_PNG_NoQuality(t *testing.T) {
	task := &model.Task{
		Mode:   model.ModeEverySecond,
		Format: "png",
	}
	args, err := BuildArgs(task, "/input/video.mp4", "/output")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	argStr := strings.Join(args, " ")
	if strings.Contains(argStr, "-q:v") {
		t.Error("PNG should not have -q:v quality param")
	}
}

func TestBuildOutputPattern(t *testing.T) {
	tests := []struct {
		naming    string
		format    string
		videoPath string
		expected  string
	}{
		{"sequential", "jpeg", "/videos/test.mp4", "frame_%04d.jpg"},
		{"sequential", "png", "/videos/test.mp4", "frame_%04d.png"},
		{"timestamp", "jpeg", "/videos/test.mp4", "%d.jpg"},
		{"video-seq", "jpeg", "/videos/test.mp4", "test_%04d.jpg"},
		{"video-time", "jpeg", "/videos/test.mp4", "test_%d.jpg"},
	}
	for _, tt := range tests {
		result := BuildOutputPattern(tt.naming, tt.format, tt.videoPath)
		if result != tt.expected {
			t.Errorf("BuildOutputPattern(%s,%s,%s) = %s, want %s",
				tt.naming, tt.format, tt.videoPath, result, tt.expected)
		}
	}
}

func TestQualityToFFmpeg(t *testing.T) {
	if q := QualityToFFmpeg(100); q != 1 {
		t.Errorf("QualityToFFmpeg(100) = %d, want 1", q)
	}
	if q := QualityToFFmpeg(1); q != 31 {
		t.Errorf("QualityToFFmpeg(1) = %d, want 31", q)
	}
	q := QualityToFFmpeg(50)
	if q < 1 || q > 31 {
		t.Errorf("QualityToFFmpeg(50) = %d, out of range [1,31]", q)
	}
}
