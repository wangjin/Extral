package recorder

import (
	"testing"
)

func TestParseAVFoundationDevices_VideoAndAudio(t *testing.T) {
	output := `[AVFoundation indev @ 0x1] AVFoundation video devices:
[AVFoundation indev @ 0x1] [0] FaceTime HD Camera (Built-in)
[AVFoundation indev @ 0x1] [1] OBS Virtual Camera
[AVFoundation indev @ 0x1] AVFoundation audio devices:
[AVFoundation indev @ 0x1] [0] MacBook Pro Microphone (Built-in)
[AVFoundation indev @ 0x1] [1] External Microphone
`

	cameras, audios := parseAVFoundationDevices(output)
	if len(cameras) != 2 {
		t.Fatalf("expected 2 cameras, got %d", len(cameras))
	}
	if cameras[0].ID != "0" || cameras[0].Name != "FaceTime HD Camera (Built-in)" {
		t.Errorf("camera[0] unexpected: %+v", cameras[0])
	}
	if cameras[1].ID != "1" || cameras[1].Name != "OBS Virtual Camera" {
		t.Errorf("camera[1] unexpected: %+v", cameras[1])
	}
	if len(audios) != 2 {
		t.Fatalf("expected 2 audio devices, got %d", len(audios))
	}
	if audios[0].ID != "0" || audios[0].Name != "MacBook Pro Microphone (Built-in)" {
		t.Errorf("audio[0] unexpected: %+v", audios[0])
	}
}

func TestParseAVFoundationDevices_Empty(t *testing.T) {
	output := `[AVFoundation indev @ 0x1] AVFoundation video devices:
[AVFoundation indev @ 0x1] AVFoundation audio devices:
`
	cameras, audios := parseAVFoundationDevices(output)
	if len(cameras) != 0 {
		t.Errorf("expected 0 cameras, got %d", len(cameras))
	}
	if len(audios) != 0 {
		t.Errorf("expected 0 audio devices, got %d", len(audios))
	}
}

func TestParseDShowDevices(t *testing.T) {
	output := `[dshow @ 0x1] "DirectShow video devices"
[dshow @ 0x1]  Alternative name "@device_pnp_..."
[dshow @ 0x1]  [Integrated Webcam]
[dshow @ 0x1]  Alternative name "@device_pnp_..."
[dshow @ 0x1] "DirectShow audio devices"
[dshow @ 0x1]  [Microphone Array (Realtek Audio)]
[dshow @ 0x1]  [External Mic]
`

	cameras, audios := parseDShowDevices(output)
	if len(cameras) != 1 {
		t.Fatalf("expected 1 camera, got %d", len(cameras))
	}
	if cameras[0].Name != "Integrated Webcam" {
		t.Errorf("camera[0] unexpected: %+v", cameras[0])
	}
	if len(audios) != 2 {
		t.Fatalf("expected 2 audio devices, got %d", len(audios))
	}
	if audios[0].Name != "Microphone Array (Realtek Audio)" {
		t.Errorf("audio[0] unexpected: %+v", audios[0])
	}
}

func TestParseSystemProfiler(t *testing.T) {
	output := `Camera:
    Camera Model:

        FaceTime HD Camera (Built-in):

          Unique ID: 0x12345678

    Camera Model:

        OBS Virtual Camera:

          Unique ID: 0x87654321
`

	cameras := parseSystemProfiler(output)
	if len(cameras) != 2 {
		t.Fatalf("expected 2 cameras, got %d", len(cameras))
	}
	if cameras[0].Name != "FaceTime HD Camera (Built-in)" {
		t.Errorf("camera[0] unexpected: %+v", cameras[0])
	}
	if cameras[1].Name != "OBS Virtual Camera" {
		t.Errorf("camera[1] unexpected: %+v", cameras[1])
	}
}
