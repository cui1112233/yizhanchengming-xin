package httpapi

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"strings"

	"qiantie/backend/internal/agentworkspace"
)

type AgentImageProcessInput struct {
	Operation   agentworkspace.ImageToolOperation
	Source      agentworkspace.Asset
	SourceBytes []byte
	Settings    map[string]any
}

type AgentImageProcessOutput struct {
	Name        string
	ContentType string
	Extension   string
	Bytes       []byte
}

type AgentImageProcessor interface {
	Process(context.Context, AgentImageProcessInput) ([]AgentImageProcessOutput, error)
}

type agentImageProcessor struct{}

func NewAgentImageProcessor() AgentImageProcessor { return &agentImageProcessor{} }

func (p *agentImageProcessor) Process(ctx context.Context, input AgentImageProcessInput) ([]AgentImageProcessOutput, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(input.SourceBytes) == 0 || input.Source.Kind != "image" || agentworkspace.ValidateImageToolSettings(input.Operation, input.Settings) != nil {
		return nil, agentworkspace.ErrInvalidImageTool
	}
	source, _, err := image.Decode(bytes.NewReader(input.SourceBytes))
	if err != nil {
		return nil, fmt.Errorf("源图片无法读取")
	}
	var images []image.Image
	switch input.Operation {
	case agentworkspace.ImageCrop:
		result, cropErr := cropAgentImage(source, imageToolNumber(input.Settings, "x"), imageToolNumber(input.Settings, "y"), imageToolNumber(input.Settings, "width"), imageToolNumber(input.Settings, "height"))
		if cropErr != nil {
			return nil, cropErr
		}
		images = []image.Image{result}
	case agentworkspace.ImageRotate:
		images = []image.Image{rotateAgentImage(source, imageToolNumber(input.Settings, "degrees"))}
	case agentworkspace.ImageMirror:
		images = []image.Image{mirrorAgentImage(source, strings.TrimSpace(fmt.Sprint(input.Settings["direction"])))}
	case agentworkspace.ImageGridSplit:
		var splitErr error
		images, splitErr = splitAgentImage(source, imageToolNumber(input.Settings, "columns"), imageToolNumber(input.Settings, "rows"))
		if splitErr != nil {
			return nil, splitErr
		}
	default:
		return nil, agentworkspace.ErrInvalidImageTool
	}
	outputs := make([]AgentImageProcessOutput, 0, len(images))
	for index, result := range images {
		var encoded bytes.Buffer
		if err := png.Encode(&encoded, result); err != nil {
			return nil, err
		}
		name := "图片编辑结果.png"
		if len(images) > 1 {
			name = fmt.Sprintf("图片拆分-%02d.png", index+1)
		}
		outputs = append(outputs, AgentImageProcessOutput{Name: name, ContentType: "image/png", Extension: "png", Bytes: encoded.Bytes()})
	}
	return outputs, nil
}

func imageToolNumber(settings map[string]any, key string) int {
	switch value := settings[key].(type) {
	case int:
		return value
	case int64:
		return int(value)
	case float64:
		return int(value)
	default:
		return 0
	}
}

func agentImageRGBA(source image.Image, bounds image.Rectangle) *image.RGBA {
	result := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(result, result.Bounds(), source, bounds.Min, draw.Src)
	return result
}

func cropAgentImage(source image.Image, x, y, width, height int) (image.Image, error) {
	bounds := source.Bounds()
	region := image.Rect(bounds.Min.X+x, bounds.Min.Y+y, bounds.Min.X+x+width, bounds.Min.Y+y+height)
	if width <= 0 || height <= 0 || !region.In(bounds) {
		return nil, fmt.Errorf("裁剪区域超出图片范围")
	}
	return agentImageRGBA(source, region), nil
}

func rotateAgentImage(source image.Image, degrees int) image.Image {
	bounds := source.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if degrees == 180 {
		result := image.NewRGBA(image.Rect(0, 0, width, height))
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				result.Set(width-1-x, height-1-y, source.At(bounds.Min.X+x, bounds.Min.Y+y))
			}
		}
		return result
	}
	result := image.NewRGBA(image.Rect(0, 0, height, width))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if degrees == 90 {
				result.Set(height-1-y, x, source.At(bounds.Min.X+x, bounds.Min.Y+y))
			} else {
				result.Set(y, width-1-x, source.At(bounds.Min.X+x, bounds.Min.Y+y))
			}
		}
	}
	return result
}

func mirrorAgentImage(source image.Image, direction string) image.Image {
	bounds := source.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	result := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			targetX, targetY := width-1-x, y
			if direction == "vertical" {
				targetX, targetY = x, height-1-y
			}
			result.Set(targetX, targetY, source.At(bounds.Min.X+x, bounds.Min.Y+y))
		}
	}
	return result
}

func splitAgentImage(source image.Image, columns, rows int) ([]image.Image, error) {
	bounds := source.Bounds()
	if columns < 1 || rows < 1 || columns > bounds.Dx() || rows > bounds.Dy() {
		return nil, fmt.Errorf("拆分网格超出图片范围")
	}
	results := make([]image.Image, 0, columns*rows)
	for row := 0; row < rows; row++ {
		top, bottom := bounds.Min.Y+(bounds.Dy()*row)/rows, bounds.Min.Y+(bounds.Dy()*(row+1))/rows
		for column := 0; column < columns; column++ {
			left, right := bounds.Min.X+(bounds.Dx()*column)/columns, bounds.Min.X+(bounds.Dx()*(column+1))/columns
			results = append(results, agentImageRGBA(source, image.Rect(left, top, right, bottom)))
		}
	}
	return results, nil
}
