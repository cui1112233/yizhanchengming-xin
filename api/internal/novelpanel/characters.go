package novelpanel

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
)

var bracketPattern = regexp.MustCompile(`[（(\[]([^()（）\[\]]{1,120})[）)\]]`)
var stripBracketPattern = regexp.MustCompile(`[（(][^()（）]*[）)]|\[[^\[\]]*\]`)

func splitForcedRoster(input string) []string {
	entries := make([]string, 0)
	var current strings.Builder
	depth := 0
	flush := func() {
		value := strings.TrimSpace(current.String())
		current.Reset()
		if value != "" {
			entries = append(entries, value)
		}
	}
	for _, r := range input {
		switch r {
		case '(', '（', '[':
			depth++
			current.WriteRune(r)
		case ')', '）', ']':
			if depth > 0 {
				depth--
			}
			current.WriteRune(r)
		case '\n', ',', '，', '、', ';', '；', '|':
			if depth == 0 {
				flush()
				continue
			}
			current.WriteRune(r)
		default:
			current.WriteRune(r)
		}
	}
	flush()
	return entries
}

func stableCharacterID(entry string, index int) string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%d\n%s", index, entry)))
	return "char_" + hex.EncodeToString(hash[:])[:20]
}

func ParseForcedRoster(input string) []CharacterCard {
	seen := map[string]struct{}{}
	cards := make([]CharacterCard, 0)
	for _, raw := range splitForcedRoster(input) {
		entry := strings.TrimSpace(raw)
		if _, ok := seen[entry]; ok {
			continue
		}
		seen[entry] = struct{}{}
		baseName := strings.TrimSpace(stripBracketPattern.ReplaceAllString(entry, ""))
		if baseName == "" {
			continue
		}
		matches := bracketPattern.FindAllStringSubmatch(entry, -1)
		hints := make([]string, 0, len(matches))
		for _, match := range matches {
			if len(match) > 1 {
				hint := strings.TrimSpace(match[1])
				if hint != "" {
					hints = append(hints, hint)
				}
			}
		}
		cards = append(cards, CharacterCard{
			ID:           stableCharacterID(entry, len(cards)+1),
			SourceEntry:  entry,
			DisplayName:  entry,
			BaseName:     baseName,
			BracketHints: hints,
			StageHint:    firstString(hints),
			Gender:       "待确认",
		})
	}
	return cards
}

func firstString(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func MergeForcedCharacters(forced []CharacterCard, submitted []CharacterCard) []CharacterCard {
	if len(forced) == 0 {
		result := make([]CharacterCard, 0, len(submitted))
		seen := map[string]struct{}{}
		for _, card := range submitted {
			card.BaseName = strings.TrimSpace(card.BaseName)
			card.DisplayName = strings.TrimSpace(card.DisplayName)
			if card.BaseName == "" {
				card.BaseName = card.DisplayName
			}
			if card.DisplayName == "" {
				card.DisplayName = card.BaseName
			}
			if card.ID == "" && card.BaseName != "" {
				card.ID = stableCharacterID(card.BaseName, len(result)+1)
			}
			if card.ID == "" {
				continue
			}
			if _, ok := seen[card.ID]; ok {
				continue
			}
			seen[card.ID] = struct{}{}
			result = append(result, card)
		}
		return result
	}

	bySource := map[string]CharacterCard{}
	byBase := map[string]CharacterCard{}
	for _, card := range submitted {
		if key := strings.TrimSpace(card.SourceEntry); key != "" {
			bySource[key] = card
		}
		if key := strings.TrimSpace(card.BaseName); key != "" {
			if _, exists := byBase[key]; !exists {
				byBase[key] = card
			}
		}
	}
	result := make([]CharacterCard, 0, len(forced))
	for _, authoritative := range forced {
		edited, ok := bySource[authoritative.SourceEntry]
		if !ok {
			edited, ok = byBase[authoritative.BaseName]
		}
		if ok {
			authoritative.Aliases = append([]string(nil), edited.Aliases...)
			if strings.TrimSpace(edited.Gender) != "" {
				authoritative.Gender = strings.TrimSpace(edited.Gender)
			}
			authoritative.Appearance = strings.TrimSpace(edited.Appearance)
			authoritative.CardNote = strings.TrimSpace(edited.CardNote)
			authoritative.AssetRefs = append([]string(nil), edited.AssetRefs...)
		}
		result = append(result, authoritative)
	}
	return result
}

func ValidateRelationships(characters []CharacterCard, relationships []CharacterRelationship) error {
	ids := make(map[string]struct{}, len(characters))
	for _, character := range characters {
		id := strings.TrimSpace(character.ID)
		if id == "" {
			return fmt.Errorf("%w: character id is required", ErrInvalid)
		}
		ids[id] = struct{}{}
	}
	seen := map[string]struct{}{}
	for _, relationship := range relationships {
		from := strings.TrimSpace(relationship.FromID)
		to := strings.TrimSpace(relationship.ToID)
		kind := strings.TrimSpace(relationship.Type)
		if from == "" || to == "" || kind == "" {
			return fmt.Errorf("%w: relationship endpoints and type are required", ErrInvalid)
		}
		if from == to {
			return fmt.Errorf("%w: self relationship is not allowed", ErrInvalid)
		}
		if _, ok := ids[from]; !ok {
			return fmt.Errorf("%w: relationship references unknown character %s", ErrInvalid, from)
		}
		if _, ok := ids[to]; !ok {
			return fmt.Errorf("%w: relationship references unknown character %s", ErrInvalid, to)
		}
		key := from + "\x00" + to + "\x00" + kind
		if _, ok := seen[key]; ok {
			return fmt.Errorf("%w: duplicate relationship", ErrInvalid)
		}
		seen[key] = struct{}{}
	}
	return nil
}
