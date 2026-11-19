package novelcreation

import (
	"encoding/json"
	"fmt"
	"strings"
)

// CM schema adapter for native incremental-state preservation. It is not a
// parser for the complete original RUNTIME_STATE_DELTA. See STATE-SPEC.md.
func mergeNativeSettlement(previous json.RawMessage, delta truthSettlementOutput, rawDelta json.RawMessage, chapter int) (truthSettlementOutput, json.RawMessage, error) {
	var old truthSettlementOutput
	if len(previous) > 0 {
		var fields map[string]json.RawMessage
		if e := json.Unmarshal(previous, &fields); e != nil || fields == nil {
			return delta, nil, &DomainError{Code: ErrPersistenceFailed, Message: "前文结算资料无法解析，已停止保存", Cause: e}
		}
		if len(fields) > 0 && fields["facts"] == nil && fields["characterChanges"] == nil && fields["timeline"] == nil && fields["foreshadowing"] == nil {
			return delta, nil, &DomainError{Code: ErrPersistenceFailed, Message: "前文结算资料格式不兼容，已停止保存"}
		}
		if e := json.Unmarshal(previous, &old); e != nil {
			return delta, nil, &DomainError{Code: ErrPersistenceFailed, Message: "前文结算条目无效，已停止保存", Cause: e}
		}
	}
	if old.Facts == nil {
		old.Facts = []settledFact{}
	}
	if old.CharacterChanges == nil {
		old.CharacterChanges = []settledCharacterChange{}
	}
	if old.Timeline == nil {
		old.Timeline = []settledTimelineEvent{}
	}
	if old.Foreshadowing == nil {
		old.Foreshadowing = []settledForeshadowing{}
	}
	if e := validateSettlement(old); e != nil {
		return delta, nil, &DomainError{Code: ErrPersistenceFailed, Message: "前文结算条目无效，已停止保存", Cause: e}
	}
	for _, value := range []truthSettlementOutput{old, delta} {
		if e := uniqueSettlementIdentifiers(value); e != nil {
			return delta, nil, &DomainError{Code: ErrInvalidOutput, Message: e.Error(), Cause: e}
		}
	}
	merged := truthSettlementOutput{
		Facts:            mergeSettlementRows(old.Facts, delta.Facts, func(v settledFact) string { return v.ID }, func(_, new settledFact) settledFact { return new }),
		CharacterChanges: mergeSettlementRows(old.CharacterChanges, delta.CharacterChanges, func(v settledCharacterChange) settledCharacterChange { return v }, func(_, new settledCharacterChange) settledCharacterChange { return new }),
		Timeline:         mergeSettlementRows(old.Timeline, delta.Timeline, func(v settledTimelineEvent) settledTimelineEvent { return v }, func(_, new settledTimelineEvent) settledTimelineEvent { return new }),
		Foreshadowing: mergeSettlementRows(old.Foreshadowing, delta.Foreshadowing, func(v settledForeshadowing) string { return v.Thread }, func(old, new settledForeshadowing) settledForeshadowing {
			if old.Status == "resolved" {
				new.Status = "resolved"
			}
			return new
		}),
	}
	if e := validateSettlement(merged); e != nil {
		return delta, nil, &DomainError{Code: ErrInvalidOutput, Message: "结算合并校验失败，已停止保存", Cause: e}
	}
	snapshot, e := json.Marshal(struct {
		truthSettlementOutput
		Delta   json.RawMessage `json:"settlement_delta"`
		Adapter string          `json:"settlement_adapter"`
		Chapter int             `json:"settlement_chapter"`
	}{merged, rawDelta, "cm_incremental_v1", chapter})
	return merged, snapshot, e
}

func uniqueSettlementIdentifiers(value truthSettlementOutput) error {
	seen := map[string]bool{}
	for _, fact := range value.Facts {
		if seen[fact.ID] {
			return fmt.Errorf("结算事实编号重复：%s，已停止保存", fact.ID)
		}
		seen[fact.ID] = true
	}
	seen = map[string]bool{}
	for _, hook := range value.Foreshadowing {
		if seen[hook.Thread] {
			return fmt.Errorf("结算伏笔重复：%s，已停止保存", strings.TrimSpace(hook.Thread))
		}
		seen[hook.Thread] = true
	}
	return nil
}

func mergeSettlementRows[T any, K comparable](old, delta []T, key func(T) K, combine func(T, T) T) []T {
	out := make([]T, 0, len(old)+len(delta))
	out = append(out, old...)
	index := map[K]int{}
	for i, v := range out {
		index[key(v)] = i
	}
	for _, v := range delta {
		k := key(v)
		if i, exists := index[k]; exists {
			out[i] = combine(out[i], v)
		} else {
			index[k] = len(out)
			out = append(out, v)
		}
	}
	return out
}
