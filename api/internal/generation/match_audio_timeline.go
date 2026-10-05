package generation

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
)

var ErrTimelineValidation = errors.New("generation: match-audio timeline validation failed")

// TimelineRepairToleranceMS is intentionally small: it permits only rounding and
// closing drift, never material timeline errors. Keep this as the single source
// of truth for deterministic timeline repair tolerance.
const TimelineRepairToleranceMS int64 = 100

type TimelineValidationResult struct {
	DurationMS int64 `json:"durationMs"`
	Repaired   bool  `json:"repaired"`
}

type timelineCard map[string]json.RawMessage

func validateAndRepairDirectorOutput(output string, targetDurationMS, shotDurationLimitMS int64) (string, TimelineValidationResult, error) {
	if targetDurationMS <= 0 || shotDurationLimitMS <= 0 {
		return "", TimelineValidationResult{}, fmt.Errorf("%w: invalid duration constraint", ErrTimelineValidation)
	}

	var document map[string]json.RawMessage
	if err := json.Unmarshal([]byte(output), &document); err != nil {
		return "", TimelineValidationResult{}, fmt.Errorf("%w: invalid JSON", ErrTimelineValidation)
	}

	cardsKey := "cards"
	cardsRaw, ok := document[cardsKey]
	if !ok {
		cardsKey = "director_cards"
		cardsRaw, ok = document[cardsKey]
	}
	if !ok {
		return "", TimelineValidationResult{}, fmt.Errorf("%w: missing cards", ErrTimelineValidation)
	}

	var cards []timelineCard
	if err := json.Unmarshal(cardsRaw, &cards); err != nil || len(cards) == 0 {
		return "", TimelineValidationResult{}, fmt.Errorf("%w: empty or invalid cards", ErrTimelineValidation)
	}

	starts := make([]int64, len(cards))
	ends := make([]int64, len(cards))
	for i, card := range cards {
		start, err := rawSecondsToMilliseconds(card["start"])
		if err != nil {
			return "", TimelineValidationResult{}, fmt.Errorf("%w: card %d invalid start", ErrTimelineValidation, i)
		}
		end, err := rawSecondsToMilliseconds(card["end"])
		if err != nil {
			return "", TimelineValidationResult{}, fmt.Errorf("%w: card %d invalid end", ErrTimelineValidation, i)
		}
		if start < 0 || end <= start {
			return "", TimelineValidationResult{}, fmt.Errorf("%w: card %d reversed or empty", ErrTimelineValidation, i)
		}
		if end-start > shotDurationLimitMS {
			return "", TimelineValidationResult{}, fmt.Errorf("%w: card %d exceeds shot duration limit", ErrTimelineValidation, i)
		}
		starts[i], ends[i] = start, end
	}

	repaired := false
	if abs64(starts[0]) > TimelineRepairToleranceMS {
		return "", TimelineValidationResult{}, fmt.Errorf("%w: first card does not start at zero", ErrTimelineValidation)
	}
	if starts[0] != 0 {
		starts[0] = 0
		repaired = true
	}

	for i := 1; i < len(cards); i++ {
		delta := starts[i] - ends[i-1]
		if abs64(delta) > TimelineRepairToleranceMS {
			kind := "gap"
			if delta < 0 {
				kind = "overlap"
			}
			return "", TimelineValidationResult{}, fmt.Errorf("%w: severe %s before card %d", ErrTimelineValidation, kind, i)
		}
		if delta != 0 {
			starts[i] = ends[i-1]
			repaired = true
			if ends[i] <= starts[i] || ends[i]-starts[i] > shotDurationLimitMS {
				return "", TimelineValidationResult{}, fmt.Errorf("%w: repair would violate card %d duration", ErrTimelineValidation, i)
			}
		}
	}

	closingDelta := targetDurationMS - ends[len(ends)-1]
	if abs64(closingDelta) > TimelineRepairToleranceMS {
		return "", TimelineValidationResult{}, fmt.Errorf("%w: total duration drift %dms exceeds tolerance", ErrTimelineValidation, closingDelta)
	}
	if closingDelta != 0 {
		last := len(ends) - 1
		ends[last] = targetDurationMS
		repaired = true
		if ends[last] <= starts[last] || ends[last]-starts[last] > shotDurationLimitMS {
			return "", TimelineValidationResult{}, fmt.Errorf("%w: closing repair violates final shot duration", ErrTimelineValidation)
		}
	}

	for i := range cards {
		cards[i]["start"] = fixedSecondsRaw(starts[i])
		cards[i]["end"] = fixedSecondsRaw(ends[i])
	}
	encodedCards, err := json.Marshal(cards)
	if err != nil {
		return "", TimelineValidationResult{}, err
	}
	document[cardsKey] = encodedCards
	encoded, err := json.Marshal(document)
	if err != nil {
		return "", TimelineValidationResult{}, err
	}
	return string(encoded), TimelineValidationResult{DurationMS: targetDurationMS, Repaired: repaired}, nil
}

func rawSecondsToMilliseconds(raw json.RawMessage) (int64, error) {
	if len(bytes.TrimSpace(raw)) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return 0, errors.New("missing time")
	}
	value, err := strconv.ParseFloat(string(bytes.TrimSpace(raw)), 64)
	if err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, errors.New("invalid time")
	}
	return int64(math.Round(value * 1000)), nil
}

func fixedSecondsRaw(milliseconds int64) json.RawMessage {
	return json.RawMessage([]byte(fmt.Sprintf("%.2f", float64(milliseconds)/1000)))
}

func abs64(value int64) int64 {
	if value < 0 {
		return -value
	}
	return value
}
