package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusPending   = "pending"
	StatusRunning   = "running"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
	StatusCancelled = "cancelled"

	ModeEveryFrame  = "every-frame"
	ModeEverySecond = "every-second"
	ModeTimeRange   = "time-range"
	ModeTotalFrames = "total-frames"
	ModeKeyframe    = "keyframe"
	ModeSceneChange = "scene-change"
	ModeTimestamps  = "timestamps"
)

type Task struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	VideoPath string    `json:"videoPath"`

	Mode   string     `json:"mode"`
	Params ModeParams `json:"params"`

	Format    string `json:"format"`
	Quality   int    `json:"quality"`
	ScaleMode string `json:"scaleMode"`
	Scale     int    `json:"scale"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	KeepRatio bool   `json:"keepRatio"`

	OutputDir string `json:"outputDir"`
	Naming    string `json:"naming"`

	Progress    float64   `json:"progress"`
	Current     int       `json:"current"`
	Total       int       `json:"total"`
	ElapsedTime int64     `json:"elapsedTime"`
	Error       string    `json:"error"`
	Logs        []string  `json:"logs"`
	CreatedAt   time.Time `json:"createdAt"`
}

type ModeParams struct {
	FrameInterval  int      `json:"frameInterval"`
	SecondInterval float64  `json:"secondInterval"`
	StartTime      string   `json:"startTime"`
	EndTime        string   `json:"endTime"`
	TotalFrames    int      `json:"totalFrames"`
	SceneThreshold float64  `json:"sceneThreshold"`
	Timestamps     []string `json:"timestamps"`
}

func NewDefaultTask(videoPath string) *Task {
	return &Task{
		ID:        uuid.New().String(),
		Status:    StatusPending,
		VideoPath: videoPath,
		Mode:      ModeEverySecond,
		Params: ModeParams{
			SecondInterval: 1,
			SceneThreshold: 0.3,
		},
		Format:    "jpeg",
		Quality:   80,
		ScaleMode: "percentage",
		Scale:     100,
		Naming:    "sequential",
		CreatedAt: time.Now(),
	}
}
