package task

import (
	"sync"

	"video-extractor/internal/model"
)

type Callback func(task *model.Task)

type Manager struct {
	mu       sync.RWMutex
	tasks    []*model.Task
	callback Callback
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) SetCallback(cb Callback) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callback = cb
}

func (m *Manager) AddTask(videoPath string) *model.Task {
	task := model.NewDefaultTask(videoPath)
	m.mu.Lock()
	m.tasks = append(m.tasks, task)
	cb := m.callback
	m.mu.Unlock()
	if cb != nil {
		cb(task)
	}
	return task
}

func (m *Manager) GetTask(id string) *model.Task {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, t := range m.tasks {
		if t.ID == id {
			return t
		}
	}
	return nil
}

func (m *Manager) ListTasks() []*model.Task {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*model.Task, len(m.tasks))
	copy(result, m.tasks)
	return result
}

func (m *Manager) UpdateTask(task *model.Task) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, t := range m.tasks {
		if t.ID == task.ID {
			m.tasks[i] = task
			break
		}
	}
	if m.callback != nil {
		m.callback(task)
	}
}

func (m *Manager) RemoveTask(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, t := range m.tasks {
		if t.ID == id {
			m.tasks = append(m.tasks[:i], m.tasks[i+1:]...)
			break
		}
	}
}

func (m *Manager) ClearTasks() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tasks = nil
}

func (m *Manager) GetPendingTasks() []*model.Task {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*model.Task
	for _, t := range m.tasks {
		if t.Status == model.StatusPending {
			result = append(result, t)
		}
	}
	return result
}
