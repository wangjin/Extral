package recorder

import "time"

// Camera represents a video capture device.
type Camera struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// AudioDevice represents an audio capture device.
type AudioDevice struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// RecordingConfig holds all parameters for a recording session.
type RecordingConfig struct {
	CameraID    string `json:"cameraId"`
	AudioDevice string `json:"audioDevice"`
	Resolution  string `json:"resolution"`
	FPS         int    `json:"fps"`
	Codec       string `json:"codec"`
	RateControl string `json:"rateControl"` // "crf" | "bitrate"
	Quality     string `json:"quality"`     // "low" | "medium" | "high" | "custom"
	CRF         int    `json:"crf"`
	Bitrate     int    `json:"bitrate"`
	Format      string `json:"format"`
	MaxDuration int    `json:"maxDuration"`
	OutputDir   string `json:"outputDir"`
	FileName    string `json:"fileName"`
}

// RecordingStatus represents the current state of a recording session.
type RecordingStatus struct {
	State    string  `json:"state"`    // "idle" | "previewing" | "recording" | "paused"
	Duration int64   `json:"duration"` // seconds
	FileSize int64   `json:"fileSize"` // bytes
	FPS      float64 `json:"fps"`
	Dropped  int     `json:"dropped"`
	Error    string  `json:"error"`
}

// QualityToCRF maps quality presets to CRF values for libx264.
func QualityToCRF(quality string) int {
	switch quality {
	case "low":
		return 28
	case "medium":
		return 23
	case "high":
		return 18
	default:
		return 23
	}
}

// DefaultRecordingConfig returns a sensible default configuration.
func DefaultRecordingConfig() RecordingConfig {
	return RecordingConfig{
		Resolution:  "1280x720",
		FPS:         30,
		Codec:       "libx264",
		RateControl: "crf",
		Quality:     "medium",
		CRF:         23,
		Format:      "mp4",
	}
}

// GenerateFileName creates a timestamp-based filename.
func GenerateFileName(format string) string {
	ext := "mp4"
	if format == "mkv" {
		ext = "mkv"
	}
	return "recording_" + time.Now().Format("2006-01-02_15-04-05") + "." + ext
}
