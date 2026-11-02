package agentworkspace

import (
	"errors"
	"fmt"
	"strings"
)

var ErrInvalidVideoTool = errors.New("invalid video tool request")

type VideoToolOperation string

const (
	VideoTrim          VideoToolOperation = "trim"
	VideoCrop          VideoToolOperation = "crop"
	VideoExtractFrames VideoToolOperation = "extract_frames"
	VideoSeparateAudio VideoToolOperation = "separate_audio"
)

func (operation VideoToolOperation) ResultKind() string {
	switch operation {
	case VideoTrim, VideoCrop:
		return "video"
	case VideoExtractFrames:
		return "image"
	case VideoSeparateAudio:
		return "audio"
	default:
		return ""
	}
}

func (operation VideoToolOperation) GenerationMode() string {
	if operation.ResultKind() == "" {
		return ""
	}
	return "video_" + string(operation)
}

func ParseVideoToolOperation(value string) (VideoToolOperation, error) {
	operation := VideoToolOperation(strings.TrimSpace(value))
	if operation.ResultKind() == "" {
		return "", ErrInvalidVideoTool
	}
	return operation, nil
}

func ValidateVideoToolSettings(operation VideoToolOperation, settings map[string]any) error {
	number := func(key string) float64 {
		switch value := settings[key].(type) {
		case float64:
			return value
		case float32:
			return float64(value)
		case int:
			return float64(value)
		case int64:
			return float64(value)
		default:
			return 0
		}
	}
	switch operation {
	case VideoTrim:
		if number("startMs") < 0 || number("endMs") <= number("startMs") {
			return ErrInvalidVideoTool
		}
	case VideoCrop:
		if number("x") < 0 || number("y") < 0 || number("width") < 1 || number("height") < 1 {
			return ErrInvalidVideoTool
		}
	case VideoExtractFrames:
		raw, ok := settings["timestampsMs"].([]any)
		if !ok || len(raw) < 1 || len(raw) > 12 {
			return ErrInvalidVideoTool
		}
		previous := float64(-1)
		for _, item := range raw {
			value, ok := item.(float64)
			if !ok || value < 0 || value <= previous {
				return ErrInvalidVideoTool
			}
			previous = value
		}
	case VideoSeparateAudio:
		return nil
	default:
		return fmt.Errorf("%w: unsupported operation", ErrInvalidVideoTool)
	}
	return nil
}
