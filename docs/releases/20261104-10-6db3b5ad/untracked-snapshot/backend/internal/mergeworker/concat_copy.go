package mergeworker

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"

	"qiantie/backend/internal/ffmpegbin"
)

// EnvConcatCopy 控制“无需重编码的容器级拼接”快速路径，默认关闭。
// 只有所有片段与目标画布参数完全一致时才会生效，否则自动回退到重编码。
const EnvConcatCopy = "QIANTIE_FFMPEG_CONCAT_COPY"

func concatCopyEnabled() bool {
	return strings.TrimSpace(os.Getenv(EnvConcatCopy)) == "1"
}

type probeStream struct {
	CodecType         string `json:"codec_type"`
	CodecName         string `json:"codec_name"`
	Width             int    `json:"width"`
	Height            int    `json:"height"`
	PixFmt            string `json:"pix_fmt"`
	SampleAspectRatio string `json:"sample_aspect_ratio"`
	RFrameRate        string `json:"r_frame_rate"`
	SampleRate        string `json:"sample_rate"`
	Channels          int    `json:"channels"`
	TimeBase          string `json:"time_base"`
}

type probeOutput struct {
	Streams []probeStream `json:"streams"`
}

func probeMediaStreams(ctx context.Context, path string) (probeStream, probeStream, bool) {
	command, _, err := ffmpegbin.Command(ctx, "ffprobe", "-v", "error", "-show_entries",
		"stream=codec_type,codec_name,width,height,pix_fmt,sample_aspect_ratio,r_frame_rate,sample_rate,channels,time_base",
		"-of", "json", path)
	if err != nil {
		return probeStream{}, probeStream{}, false
	}
	output, err := command.Output()
	if err != nil {
		return probeStream{}, probeStream{}, false
	}
	var probe probeOutput
	if json.Unmarshal(output, &probe) != nil {
		return probeStream{}, probeStream{}, false
	}
	var video, audio probeStream
	for _, stream := range probe.Streams {
		switch stream.CodecType {
		case "video":
			if video.CodecName == "" {
				video = stream
			}
		case "audio":
			if audio.CodecName == "" {
				audio = stream
			}
		}
	}
	return video, audio, video.CodecName != "" && audio.CodecName != ""
}

func sameAspect(value string) bool {
	value = strings.TrimSpace(value)
	return value == "" || value == "1:1" || value == "0:1"
}

// concatCopyCompatible 校验所有片段是否能安全地直接拷贝拼接。
// 任一参数不一致都返回 false，从而回退到重编码路径。
func concatCopyCompatible(clips []probeStream, width, height int) bool {
	if len(clips) == 0 {
		return false
	}
	first := clips[0]
	for _, clip := range clips {
		if clip.CodecName != "h264" || clip.Width != width || clip.Height != height {
			return false
		}
		if clip.PixFmt != "yuv420p" || !sameAspect(clip.SampleAspectRatio) {
			return false
		}
		if clip.RFrameRate != "30/1" || clip.TimeBase != first.TimeBase {
			return false
		}
		if clip.SampleRate != "48000" || clip.Channels != 2 {
			return false
		}
	}
	return true
}

// tryConcatCopy 在开关开启且参数完全一致时直接拷贝拼接；失败时返回 false 由调用方回退重编码。
func (r *FFmpegRunner) tryConcatCopy(ctx context.Context, manifestPath, outputPath string, width, height int) bool {
	paths, err := concatInputs(manifestPath)
	if err != nil || len(paths) == 0 {
		return false
	}
	videos := make([]probeStream, 0, len(paths))
	audios := make([]probeStream, 0, len(paths))
	for _, path := range paths {
		video, audio, ok := probeMediaStreams(ctx, path)
		if !ok {
			return false
		}
		videos = append(videos, video)
		audios = append(audios, audio)
	}
	if !concatCopyCompatible(videos, width, height) {
		return false
	}
	audioFirst := audios[0]
	for _, audio := range audios {
		if audio.CodecName != "aac" || audio.SampleRate != audioFirst.SampleRate || audio.Channels != audioFirst.Channels {
			return false
		}
	}
	log.Printf("merge: 使用容器级拼接（不重编码），片段数=%d 画布=%dx%d", len(paths), width, height)
	args := []string{"-hide_banner", "-loglevel", "error", "-y", "-f", "concat", "-safe", "0", "-i", manifestPath, "-c", "copy", "-movflags", "+faststart", outputPath}
	if err := r.run(ctx, args...); err != nil {
		log.Printf("merge: 容器级拼接失败，回退重编码：%v", err)
		return false
	}
	return true
}
