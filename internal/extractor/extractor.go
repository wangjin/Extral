package extractor

import (
	"fmt"
	"path/filepath"
	"strings"

	"video-extractor/internal/model"
)

func BuildArgs(task *model.Task, inputPath string, outputDir string) ([]string, error) {
	var args []string

	if task.Mode == model.ModeTimeRange {
		if task.Params.StartTime != "" {
			args = append(args, "-ss", task.Params.StartTime)
		}
		if task.Params.EndTime != "" {
			args = append(args, "-to", task.Params.EndTime)
		}
	}

	args = append(args, "-i", inputPath)

	vf := buildVideoFilter(task)
	if vf != "" {
		args = append(args, "-vf", vf)
	}

	needsVFR := task.Mode == model.ModeEveryFrame || task.Mode == model.ModeKeyframe ||
		task.Mode == model.ModeSceneChange || task.Mode == model.ModeTotalFrames
	if needsVFR {
		args = append(args, "-vsync", "vfr")
	}

	if task.Mode == model.ModeTotalFrames {
		args = append(args, "-frames:v", fmt.Sprintf("%d", task.Params.TotalFrames))
	}

	if task.Format == "jpeg" {
		args = append(args, "-q:v", fmt.Sprintf("%d", QualityToFFmpeg(task.Quality)))
	}

	pattern := BuildOutputPattern(task.Naming, task.Format, task.VideoPath)
	args = append(args, filepath.Join(outputDir, pattern))

	return args, nil
}

func buildVideoFilter(task *model.Task) string {
	var filters []string

	switch task.Mode {
	case model.ModeEveryFrame:
		filters = append(filters, fmt.Sprintf("select='not(mod(n\\,%d))'", task.Params.FrameInterval))
	case model.ModeEverySecond:
		filters = append(filters, fmt.Sprintf("fps=1/%g", task.Params.SecondInterval))
	case model.ModeTotalFrames:
		filters = append(filters, fmt.Sprintf(
			"select='not(mod(n\\,round(n/%d)))'", task.Params.TotalFrames,
		))
	case model.ModeKeyframe:
		filters = append(filters, "select='eq(pict_type\\,I)'")
	case model.ModeSceneChange:
		filters = append(filters, fmt.Sprintf("select='gt(scene\\,%g)'", task.Params.SceneThreshold))
	case model.ModeTimeRange:
	}

	switch task.ScaleMode {
	case "percentage":
		if task.Scale != 100 && task.Scale > 0 {
			ratio := float64(task.Scale) / 100.0
			filters = append(filters, fmt.Sprintf("scale=iw*%.2f:ih*%.2f", ratio, ratio))
		}
	case "custom":
		if task.Width > 0 {
			h := fmt.Sprintf("%d", task.Height)
			if task.KeepRatio {
				h = "-1"
			}
			filters = append(filters, fmt.Sprintf("scale=%d:%s", task.Width, h))
		}
	}

	if len(filters) == 0 {
		return ""
	}
	return strings.Join(filters, ",")
}

func BuildOutputPattern(naming string, format string, videoPath string) string {
	ext := "jpg"
	if format == "png" {
		ext = "png"
	}
	videoName := strings.TrimSuffix(filepath.Base(videoPath), filepath.Ext(videoPath))

	switch naming {
	case "sequential":
		return fmt.Sprintf("frame_%%04d.%s", ext)
	case "timestamp":
		return fmt.Sprintf("%%d.%s", ext)
	case "video-seq":
		return fmt.Sprintf("%s_%%04d.%s", videoName, ext)
	case "video-time":
		return fmt.Sprintf("%s_%%d.%s", videoName, ext)
	default:
		return fmt.Sprintf("frame_%%04d.%s", ext)
	}
}

func QualityToFFmpeg(quality int) int {
	if quality <= 0 {
		quality = 1
	}
	if quality > 100 {
		quality = 100
	}
	return 32 - (quality*31+99)/100
}

func BuildTimestampArgs(task *model.Task, outputDir string) [][]string {
	var commands [][]string
	ext := "jpg"
	if task.Format == "png" {
		ext = "png"
	}

	videoName := strings.TrimSuffix(filepath.Base(task.VideoPath), filepath.Ext(task.VideoPath))

	for i, ts := range task.Params.Timestamps {
		filename := fmt.Sprintf("%s_%04d.%s", videoName, i+1, ext)
		args := []string{
			"-ss", ts,
			"-i", task.VideoPath,
			"-frames:v", "1",
		}
		if task.Format == "jpeg" {
			args = append(args, "-q:v", fmt.Sprintf("%d", QualityToFFmpeg(task.Quality)))
		}
		args = append(args, filepath.Join(outputDir, filename))
		commands = append(commands, args)
	}
	return commands
}
