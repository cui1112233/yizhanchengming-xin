package batchfactoryv11

import (
	"context"
	"encoding/json"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

type publicDraftContext struct{}

// Kind names are protocol identifiers. Fold presentation and separator variants
// before comparison, including variants that a database collation equates.
func serverPublishDraft(kind string) bool {
	name := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return unicode.ToLower(r)
		}
		return -1
	}, norm.NFKD.String(kind))
	return strings.HasPrefix(name, "n8comicpublish") || strings.HasPrefix(name, "n8textpublish") || strings.HasPrefix(name, "collectionpublish")
}

func privateReceiptData(content string) bool {
	var v any
	if json.Unmarshal([]byte(content), &v) != nil {
		return false
	}
	var contains func(any) bool
	contains = func(v any) bool {
		switch x := v.(type) {
		case map[string]any:
			for k, child := range x {
				switch strings.ToLower(k) {
				case "_receiptpath", "_imagesignatures", "_dispatchid", "_operationid":
					return true
				}
				if contains(child) {
					return true
				}
			}
		case []any:
			for _, child := range x {
				if contains(child) {
					return true
				}
			}
		}
		return false
	}
	return contains(v)
}

func GetPublicDraft(ctx context.Context, store Store, owner, key, kind, scope string) (Draft, error) {
	if serverPublishDraft(kind) {
		return Draft{}, ErrNotFound
	}
	d, err := store.GetDraft(ctx, owner, key, kind, scope)
	if err != nil {
		return Draft{}, err
	}
	if serverPublishDraft(d.Kind) || privateReceiptData(d.Content) {
		return Draft{}, ErrNotFound
	}
	return d, nil
}

func SavePublicDraft(ctx context.Context, store Store, owner string, d Draft) (Draft, error) {
	if serverPublishDraft(d.Kind) || strings.EqualFold(strings.TrimSpace(d.Kind), AssetPromptProofKind) || privateReceiptData(d.Content) {
		return Draft{}, ErrInvalid
	}
	return store.SaveDraft(context.WithValue(ctx, publicDraftContext{}, true), owner, d)
}
