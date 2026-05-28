package task

import (
	"testing"

	"video-extractor/internal/model"
)

func TestManager_AddTask(t *testing.T) {
	m := NewManager()
	task := m.AddTask("/path/to/video.mp4")
	if task.ID == "" {
		t.Error("task should have an ID")
	}
	if task.Status != model.StatusPending {
		t.Error("new task should be pending")
	}
	if len(m.ListTasks()) != 1 {
		t.Error("should have 1 task")
	}
}

func TestManager_GetTask(t *testing.T) {
	m := NewManager()
	task := m.AddTask("/path/to/video.mp4")
	found := m.GetTask(task.ID)
	if found == nil {
		t.Fatal("task should be found")
	}
	if found.ID != task.ID {
		t.Error("ID mismatch")
	}
}

func TestManager_GetTask_NotFound(t *testing.T) {
	m := NewManager()
	found := m.GetTask("nonexistent")
	if found != nil {
		t.Error("should return nil for nonexistent task")
	}
}

func TestManager_UpdateTask(t *testing.T) {
	m := NewManager()
	task := m.AddTask("/path/to/video.mp4")
	task.Mode = model.ModeKeyframe
	task.Quality = 95
	m.UpdateTask(task)

	updated := m.GetTask(task.ID)
	if updated.Mode != model.ModeKeyframe {
		t.Errorf("expected mode %s, got %s", model.ModeKeyframe, updated.Mode)
	}
	if updated.Quality != 95 {
		t.Errorf("expected quality 95, got %d", updated.Quality)
	}
}

func TestManager_RemoveTask(t *testing.T) {
	m := NewManager()
	task := m.AddTask("/path/to/video.mp4")
	m.RemoveTask(task.ID)
	if len(m.ListTasks()) != 0 {
		t.Error("task list should be empty after removal")
	}
	if m.GetTask(task.ID) != nil {
		t.Error("removed task should not be found")
	}
}

func TestManager_ClearTasks(t *testing.T) {
	m := NewManager()
	m.AddTask("/video1.mp4")
	m.AddTask("/video2.mp4")
	m.ClearTasks()
	if len(m.ListTasks()) != 0 {
		t.Error("task list should be empty after clear")
	}
}

func TestManager_GetPendingTasks(t *testing.T) {
	m := NewManager()
	t1 := m.AddTask("/video1.mp4")
	t2 := m.AddTask("/video2.mp4")
	t3 := m.AddTask("/video3.mp4")
	t1.Status = model.StatusRunning
	m.UpdateTask(t1)
	t3.Status = model.StatusCompleted
	m.UpdateTask(t3)

	pending := m.GetPendingTasks()
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending task, got %d", len(pending))
	}
	if pending[0].ID != t2.ID {
		t.Error("pending task should be t2")
	}
}

func TestManager_CallbackFired(t *testing.T) {
	m := NewManager()
	var callbackTask *model.Task
	m.SetCallback(func(task *model.Task) {
		callbackTask = task
	})
	task := m.AddTask("/video.mp4")
	if callbackTask == nil || callbackTask.ID != task.ID {
		t.Error("callback should have been fired with the new task")
	}
}
