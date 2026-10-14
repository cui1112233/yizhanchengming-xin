package batchfactoryv11

import "strings"

// existingAssetRow is the storage-neutral view of a currently saved asset.
type existingAssetRow struct {
	ID     string
	Kind   string
	Name   string
	Source string
}

// directorAssetReconcile is the write plan for one director run.
type directorAssetReconcile struct {
	// MatchedIDs are existing director rows this run reuses. They are frozen:
	// the AI's reworded descriptions must never overwrite the saved prompts.
	MatchedIDs []string
	Inserts    []BookAsset // brand-new director assets
	Deletes    []string    // stale director asset IDs, safe to remove
}

// splitAssetName splits a trailing parenthesised note from the name, so that
// "前女友（林薇）" is recognised as the same person as "前女友".
// Only one trailing group using full-width or half-width parentheses is
// stripped; the core keeps any earlier groups.
func splitAssetName(name string) (core, note string) {
	trimmed := strings.TrimSpace(name)
	for _, pair := range [][2]rune{{'（', '）'}, {'(', ')'}} {
		open, close := pair[0], pair[1]
		lastOpen := stringsLastIndexRune(trimmed, open)
		if lastOpen <= 0 {
			continue
		}
		if !strings.HasSuffix(trimmed, string(close)) {
			continue
		}
		inner := trimmed[lastOpen+len(string(open)) : len(trimmed)-len(string(close))]
		if strings.ContainsRune(inner, open) || strings.ContainsRune(inner, close) {
			continue
		}
		return strings.TrimSpace(trimmed[:lastOpen]), strings.TrimSpace(inner)
	}
	return trimmed, ""
}

// stringsLastIndexRune keeps splitAssetName readable without importing utf8
// index math at call sites.
func stringsLastIndexRune(value string, r rune) int {
	return strings.LastIndex(value, string(r))
}

// addCoreNameLookups also indexes saved rows by their core name (parenthesised
// note stripped), so a shot whose text uses a renamed variant (e.g. AI writes
// 前女友（林薇） while the frozen row is 前女友) still binds to the same asset.
// Exact-name keys are never overwritten; an ambiguous core (two different rows
// share it) is left unset rather than guessing the wrong person.
func addCoreNameLookups(ids map[string]string, kind string, rows [][2]string) {
	coreIDs := map[string]string{}
	for _, row := range rows {
		name, id := row[0], row[1]
		core, _ := splitAssetName(name)
		key := kind + "\x00" + core
		if existing, exists := coreIDs[key]; exists && existing != id {
			coreIDs[key] = ""
			continue
		}
		if _, exists := coreIDs[key]; !exists {
			coreIDs[key] = id
		}
	}
	for key, id := range coreIDs {
		if id == "" {
			continue
		}
		if _, exists := ids[key]; !exists {
			ids[key] = id
		}
	}
}

// reconcileDirectorAssets matches the freshly extracted director assets onto
// the rows already saved for the book, then decides what to keep, insert and
// delete. Matching rules:
//  1. same kind + exact full name;
//  2. otherwise same core name (parenthesised note ignored), but only when at
//     least one side has no note and the match is unique on both sides.
//
// Matched rows are frozen: the saved name, prompt and preset stay as they are,
// even if the AI reworded them in this run. Asset extraction is a merge, not a
// destructive replacement: stale director rows remain available when a model
// response omits them, while manual rows are never changed or removed.
func reconcileDirectorAssets(existing []existingAssetRow, desired []BookAsset, hasImage map[string]bool) directorAssetReconcile {
	existingMatch := make([]int, len(existing)) // -> desired index
	desiredMatch := make([]int, len(desired))   // -> existing index
	for i := range existingMatch {
		existingMatch[i] = -1
	}
	for i := range desiredMatch {
		desiredMatch[i] = -1
	}

	// Pass 1: exact kind + name.
	byExact := map[string][]int{}
	for i, row := range existing {
		key := row.Kind + "\x00" + row.Name
		byExact[key] = append(byExact[key], i)
	}
	for di, asset := range desired {
		for _, ei := range byExact[asset.Kind+"\x00"+asset.Name] {
			if existingMatch[ei] == -1 {
				existingMatch[ei] = di
				desiredMatch[di] = ei
				break
			}
		}
	}

	// Pass 2: core name with the empty-note rule and uniqueness on both sides.
	core, note := make([]string, len(existing)), make([]string, len(existing))
	dCore, dNote := make([]string, len(desired)), make([]string, len(desired))
	for i, row := range existing {
		core[i], note[i] = splitAssetName(row.Name)
	}
	for i, asset := range desired {
		dCore[i], dNote[i] = splitAssetName(asset.Name)
	}
	for di := range desired {
		if desiredMatch[di] != -1 {
			continue
		}
		// Desired-side ambiguity: another unmatched desired shares the core.
		sameCoreDesired := 0
		for other := range desired {
			if other != di && desiredMatch[other] == -1 && desired[other].Kind == desired[di].Kind && dCore[other] == dCore[di] {
				sameCoreDesired++
			}
		}
		if sameCoreDesired > 0 {
			continue
		}
		candidates := []int{}
		for ei, row := range existing {
			if existingMatch[ei] != -1 || row.Kind != desired[di].Kind || core[ei] != dCore[di] {
				continue
			}
			// At least one side must carry no note. Two different notes (e.g.
			// 儿子（小冰山） vs 儿子（小火山）) must never be merged.
			if note[ei] != "" && dNote[di] != "" {
				continue
			}
			candidates = append(candidates, ei)
		}
		if len(candidates) == 1 {
			ei := candidates[0]
			existingMatch[ei] = di
			desiredMatch[di] = ei
		}
	}

	plan := directorAssetReconcile{}
	for di, asset := range desired {
		ei := desiredMatch[di]
		if ei == -1 {
			plan.Inserts = append(plan.Inserts, asset)
			continue
		}
		// Matched existing row (manual or director): freeze it. The saved
		// prompt wins verbatim — no rename, no prompt overwrite, no revision
		// bump — and no duplicate is inserted.
		plan.MatchedIDs = append(plan.MatchedIDs, existing[ei].ID)
	}
	// Do not delete unmatched rows here. A partial AI response must not remove
	// a previously extracted prop, scene, or character and break existing @
	// references. Explicit user deletion remains the only deletion path.
	return plan
}
