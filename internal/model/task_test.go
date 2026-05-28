package model

import "testing"

func TestTaskStatusConstants(t *testing.T) {
	statuses := []string{StatusPending, StatusRunning, StatusCompleted, StatusFailed, StatusCancelled}
	for _, s := range statuses {
		if s == "" {
			t.Error("status constant should not be empty")
		}
	}
}

func TestModeConstants(t *testing.T) {
	modes := []string{
		ModeEveryFrame, ModeEverySecond, ModeTimeRange,
		ModeTotalFrames, ModeKeyframe, ModeSceneChange, ModeTimestamps,
	}
	for _, m := range modes {
		if m == "" {
			t.Error("mode constant should not be empty")
		}
	}
}

func TestNewDefaultTask(t *testing.T) {
	task := NewDefaultTask("/path/to/video.mp4")
	if task.ID == "" {
		t.Error("task ID should not be empty")
	}
	if task.Status != StatusPending {
		t.Errorf("expected status %s, got %s", StatusPending, task.Status)
	}
	if task.VideoPath != "/path/to/video.mp4" {
		t.Error("VideoPath mismatch")
	}
	if task.Mode != ModeEverySecond {
		t.Errorf("expected mode %s, got %s", ModeEverySecond, task.Mode)
	}
	if task.Format != "jpeg" {
		t.Error("default format should be jpeg")
	}
	if task.Quality != 80 {
		t.Error("default quality should be 80")
	}
	if task.ScaleMode != "percentage" {
		t.Error("default ScaleMode should be percentage")
	}
	if task.Scale != 100 {
		t.Error("default scale should be 100")
	}
	if task.Naming != "sequential" {
		t.Error("default naming should be sequential")
	}
	if task.Params.SceneThreshold != 0.3 {
		t.Error("default scene threshold should be 0.3")
	}
}
