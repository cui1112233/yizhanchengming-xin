package agentworkspace

import (
	"errors"
	"strings"
)

var ErrInvalidImageTool = errors.New("invalid image tool request")

type ImageToolOperation string

const (
	ImageCrop      ImageToolOperation = "crop"
	ImageRotate    ImageToolOperation = "rotate"
	ImageMirror    ImageToolOperation = "mirror"
	ImageGridSplit ImageToolOperation = "grid_split"
)

func (operation ImageToolOperation) Valid() bool {
	switch operation {
	case ImageCrop, ImageRotate, ImageMirror, ImageGridSplit:
		return true
	default:
		return false
	}
}

func (operation ImageToolOperation) GenerationMode() string {
	if !operation.Valid() {
		return ""
	}
	return "image_" + string(operation)
}

func ParseImageToolOperation(value string) (ImageToolOperation, error) {
	operation := ImageToolOperation(strings.TrimSpace(value))
	if !operation.Valid() {
		return "", ErrInvalidImageTool
	}
	return operation, nil
}

func ValidateImageToolSettings(operation ImageToolOperation, settings map[string]any) error {
	if !operation.Valid() {
		return ErrInvalidImageTool
	}
	integer := func(key string) int {
		value := settings[key]
		switch typed := value.(type) {
		case int:
			return typed
		case int64:
			return int(typed)
		case float64:
			return int(typed)
		default:
			return 0
		}
	}
	switch operation {
	case ImageCrop:
		if integer("width") <= 0 || integer("height") <= 0 || integer("x") < 0 || integer("y") < 0 {
			return ErrInvalidImageTool
		}
	case ImageRotate:
		degrees := integer("degrees")
		if degrees != 90 && degrees != 180 && degrees != 270 {
			return ErrInvalidImageTool
		}
	case ImageMirror:
		direction, _ := settings["direction"].(string)
		if direction != "horizontal" && direction != "vertical" {
			return ErrInvalidImageTool
		}
	case ImageGridSplit:
		if integer("columns") < 1 || integer("columns") > 8 || integer("rows") < 1 || integer("rows") > 8 {
			return ErrInvalidImageTool
		}
	}
	return nil
}
