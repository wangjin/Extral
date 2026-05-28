package video

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type Info struct {
	Width       int     `json:"width"`
	Height      int     `json:"height"`
	FPS         float64 `json:"fps"`
	Duration    float64 `json:"duration"`
	TotalFrames int     `json:"totalFrames"`
}

type ffprobeOutput struct {
	Streams []ffprobeStream `json:"streams"`
	Format  ffprobeFormat   `json:"format"`
}

type ffprobeStream struct {
	CodecType  string `json:"codec_type"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	RFrameRate string `json:"r_frame_rate"`
	NBFrames   string `json:"nb_frames"`
	Duration   string `json:"duration"`
}

type ffprobeFormat struct {
	Duration string `json:"duration"`
}

func ParseVideoInfo(data []byte) (*Info, error) {
	var output ffprobeOutput
	if err := json.Unmarshal(data, &output); err != nil {
		return nil, fmt.Errorf("parse ffprobe output: %w", err)
	}

	var videoStream *ffprobeStream
	for i := range output.Streams {
		if output.Streams[i].CodecType == "video" {
			videoStream = &output.Streams[i]
			break
		}
	}
	if videoStream == nil {
		return nil, fmt.Errorf("no video stream found")
	}

	info := &Info{
		Width:  videoStream.Width,
		Height: videoStream.Height,
		FPS:    parseFrameRate(videoStream.RFrameRate),
	}

	if videoStream.Duration != "" {
		info.Duration, _ = strconv.ParseFloat(videoStream.Duration, 64)
	} else if output.Format.Duration != "" {
		info.Duration, _ = strconv.ParseFloat(output.Format.Duration, 64)
	}

	if videoStream.NBFrames != "" {
		info.TotalFrames, _ = strconv.Atoi(videoStream.NBFrames)
	}
	if info.TotalFrames == 0 && info.Duration > 0 && info.FPS > 0 {
		info.TotalFrames = int(info.Duration * info.FPS)
	}

	return info, nil
}

func parseFrameRate(rate string) float64 {
	parts := strings.Split(rate, "/")
	if len(parts) != 2 {
		f, _ := strconv.ParseFloat(rate, 64)
		return f
	}
	num, _ := strconv.ParseFloat(parts[0], 64)
	den, _ := strconv.ParseFloat(parts[1], 64)
	if den == 0 {
		return 0
	}
	return num / den
}
