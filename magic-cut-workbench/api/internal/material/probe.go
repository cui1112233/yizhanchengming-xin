package material

import (
	"context"
	"fmt"
	"os/exec"
)

type FFprobe struct{ Binary string }

func (p FFprobe) Probe(ctx context.Context, path string) (Metadata, error) {
	binary := p.Binary
	if binary == "" {
		binary = "ffprobe"
	}
	output, err := exec.CommandContext(ctx, binary, "-v", "error", "-print_format", "json", "-show_format", "-show_streams", path).Output()
	if err != nil {
		return Metadata{}, fmt.Errorf("ffprobe media: %w", err)
	}
	metadata, err := ParseProbeJSON(output)
	if err != nil {
		return Metadata{}, fmt.Errorf("parse ffprobe output: %w", err)
	}
	return metadata, nil
}
