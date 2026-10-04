package agent

import (
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

type toolRowFake struct {
	values []any
	err error
}
func (r toolRowFake) Scan(dest ...any) error {
	if r.err != nil { return r.err }
	for index, value := range r.values {
		switch target := dest[index].(type) {
		case *string: *target = value.(string)
		case *[]byte:
			if value == nil { *target = nil } else { *target = append([]byte(nil), value.([]byte)...) }
		case *time.Time: *target = value.(time.Time)
		default: return errors.New("unsupported scan target")
		}
	}
	return nil
}

func TestScanToolCallLoadsExecutionState(t *testing.T) {
	now := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
	item, err := scanToolCall(toolRowFake{values: []any{
		"tool_1", "thread_1", "msg_1", "owner", ToolImageGenerate, ToolProposed,
		[]byte(`{"prompt":"雨夜"}`), nil, now, now,
	}})
	if err != nil { t.Fatal(err) }
	if item.ID != "tool_1" || item.Owner != "owner" || item.ToolName != ToolImageGenerate || item.Status != ToolProposed { t.Fatalf("item=%#v", item) }
	if !json.Valid(item.Arguments) { t.Fatalf("arguments=%q", item.Arguments) }
}

func TestScanToolCallMapsMissingRow(t *testing.T) {
	if _, err := scanToolCall(toolRowFake{err: sql.ErrNoRows}); !errors.Is(err, ErrNotFound) { t.Fatalf("err=%v", err) }
}
