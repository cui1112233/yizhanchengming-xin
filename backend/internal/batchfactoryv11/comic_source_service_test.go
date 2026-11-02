package batchfactoryv11

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
)

type failingComicSegmentationProvider struct{}

func (failingComicSegmentationProvider) Complete(context.Context, TextCompletionRequest) (string, error) {
	return "", errors.New("provider unavailable")
}

func TestComicSourceSegmentationUsesValidModelBoundaryIDs(t *testing.T) {
	store, batch, source := comicSourceStoreFixture(t)
	boundaries := EnumerateComicSourceBoundaries(source)
	fallback, err := FallbackComicSourceSegments(source, boundaries)
	if err != nil {
		t.Fatal(err)
	}
	ids := make([]string, 0, len(fallback))
	for _, segment := range fallback {
		for index, boundary := range boundaries {
			if boundary.Offset == segment.EndOffset {
				ids = append(ids, fmt.Sprintf("b%d", index+1))
				break
			}
		}
	}
	payload, _ := json.Marshal(map[string]any{"boundaryIds": ids})
	provider := &queuedDirectorProvider{values: []string{string(payload)}}
	service := DirectorService{Store: store, Provider: provider}
	result, err := service.RunComicSourceSegmentation(context.Background(), "alice", batch.ID, batch.Books[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(provider.calls) != 1 || !strings.Contains(provider.calls[0].UserPrompt, "b1") {
		t.Fatalf("model did not receive numbered safe boundaries: %#v", provider.calls)
	}
	if len(result.Segments) != len(fallback) || result.Segments[0].Reason != "model_selected_boundary" {
		t.Fatalf("valid model selection was not used: %#v", result.Segments)
	}
}

func TestComicSourceSegmentationFallsBackWhenProviderFails(t *testing.T) {
	store, batch, source := comicSourceStoreFixture(t)
	service := DirectorService{Store: store, Provider: failingComicSegmentationProvider{}}
	result, err := service.RunComicSourceSegmentation(context.Background(), "alice", batch.ID, batch.Books[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	expected, _ := FallbackComicSourceSegments(source, EnumerateComicSourceBoundaries(source))
	if len(result.Segments) != len(expected) || result.Segments[0].Reason == "model_selected_boundary" {
		t.Fatalf("provider failure did not use deterministic fallback: %#v", result.Segments)
	}
}

func TestComicSourceSegmentationFallsBackOnUnsafeOrMalformedSelection(t *testing.T) {
	for _, response := range []string{`not json`, `{"boundaryIds":["b999"]}`, `{"boundaryIds":["b1"]}`} {
		t.Run(response, func(t *testing.T) {
			store, batch, source := comicSourceStoreFixture(t)
			service := DirectorService{Store: store, Provider: &queuedDirectorProvider{values: []string{response}}}
			result, err := service.RunComicSourceSegmentation(context.Background(), "alice", batch.ID, batch.Books[0].ID)
			if err != nil {
				t.Fatal(err)
			}
			expected, _ := FallbackComicSourceSegments(source, EnumerateComicSourceBoundaries(source))
			if len(result.Segments) != len(expected) {
				t.Fatalf("expected fallback with %d pages, got %d", len(expected), len(result.Segments))
			}
		})
	}
}
