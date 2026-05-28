package video

import "testing"

func TestParseVideoInfo_ValidFFprobeJSON(t *testing.T) {
	jsonOutput := `{
		"streams": [
			{
				"codec_type": "video",
				"width": 1920,
				"height": 1080,
				"r_frame_rate": "30/1",
				"nb_frames": "900",
				"duration": "30.000000"
			}
		],
		"format": {
			"duration": "30.000000"
		}
	}`

	info, err := ParseVideoInfo([]byte(jsonOutput))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Width != 1920 {
		t.Errorf("expected width 1920, got %d", info.Width)
	}
	if info.Height != 1080 {
		t.Errorf("expected height 1080, got %d", info.Height)
	}
	if info.FPS != 30 {
		t.Errorf("expected fps 30, got %f", info.FPS)
	}
	if info.Duration != 30.0 {
		t.Errorf("expected duration 30, got %f", info.Duration)
	}
	if info.TotalFrames != 900 {
		t.Errorf("expected total frames 900, got %d", info.TotalFrames)
	}
}

func TestParseVideoInfo_NoVideoStream(t *testing.T) {
	jsonOutput := `{"streams": [{"codec_type": "audio"}], "format": {"duration": "10"}}`
	_, err := ParseVideoInfo([]byte(jsonOutput))
	if err == nil {
		t.Error("expected error for no video stream")
	}
}

func TestParseVideoInfo_InvalidJSON(t *testing.T) {
	_, err := ParseVideoInfo([]byte("not json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestParseVideoInfo_MissingNbFrames_CalculatesFromDuration(t *testing.T) {
	jsonOutput := `{
		"streams": [
			{
				"codec_type": "video",
				"width": 1280,
				"height": 720,
				"r_frame_rate": "24/1",
				"duration": "60.0"
			}
		],
		"format": {"duration": "60.0"}
	}`

	info, err := ParseVideoInfo([]byte(jsonOutput))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.TotalFrames != 1440 {
		t.Errorf("expected 1440 frames (60*24), got %d", info.TotalFrames)
	}
}

func TestParseFrameRate(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"30/1", 30},
		{"24000/1001", 23.976023976023978},
		{"25/1", 25},
		{"60/1", 60},
	}
	for _, tt := range tests {
		result := parseFrameRate(tt.input)
		diff := result - tt.expected
		if diff < -0.01 || diff > 0.01 {
			t.Errorf("parseFrameRate(%s) = %f, want %f", tt.input, result, tt.expected)
		}
	}
}
