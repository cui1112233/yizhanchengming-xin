package batchfactorysensitive

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"qiantie/backend/internal/batchfactoryv11"
)

type Mode string

const (
	ModeAI     Mode = "ai"
	ModeSystem Mode = "system"
	ModeOff    Mode = "off"
)

// ResolveMode prefers an explicit valid mode and supports batches saved with the legacy AI switch.
func ResolveMode(settings batchfactoryv11.SettingsPatch) Mode {
	var mode Mode
	if json.Unmarshal(settings["sensitiveMode"], &mode) == nil {
		switch mode {
		case ModeAI, ModeSystem, ModeOff:
			return mode
		}
	}

	var enabled bool
	if json.Unmarshal(settings["sensitiveAIEnabled"], &enabled) == nil && enabled {
		return ModeAI
	}
	return ModeSystem
}

type Input struct {
	Mode                             Mode
	Owner, BookID, Text, TextModelID string
}

type Change struct {
	Keyword string
	Before  string
	After   string
}

type Result struct {
	OriginalText, ProductionText string
	Mode                         Mode
	Changes                      []Change
}

type RuleProcessor interface {
	ProcessRules(context.Context, Input) (string, []Change, error)
}

type AIProcessor interface {
	ProcessAI(context.Context, Input) (string, []Change, error)
}

type Processor struct {
	Rules RuleProcessor
	AI    AIProcessor
}

func (p Processor) Process(ctx context.Context, input Input) (Result, error) {
	result := Result{OriginalText: input.Text, Mode: input.Mode}
	var err error

	switch input.Mode {
	case ModeOff:
		result.ProductionText = input.Text
	case ModeSystem:
		if p.Rules == nil {
			return Result{}, errors.New("sensitive rules processor is not configured")
		}
		result.ProductionText, result.Changes, err = p.Rules.ProcessRules(ctx, input)
	case ModeAI:
		if p.AI == nil {
			return Result{}, errors.New("sensitive AI processor is not configured")
		}
		result.ProductionText, result.Changes, err = p.AI.ProcessAI(ctx, input)
	default:
		return Result{}, fmt.Errorf("invalid sensitive processing mode %q", input.Mode)
	}
	if err != nil {
		return Result{}, err
	}
	return result, nil
}
