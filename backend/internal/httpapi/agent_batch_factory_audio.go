package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
)

type batchFactoryProductionAudioReader struct {
	batches agentBatchFactoryBridgeStore
	files   *localartifact.Store
}

func newBatchFactoryProductionAudioReader(batches agentBatchFactoryBridgeStore, files *localartifact.Store) agentworkspace.BatchFactoryAudioReader {
	if batches == nil || files == nil {
		return nil
	}
	return &batchFactoryProductionAudioReader{batches: batches, files: files}
}

func (r *batchFactoryProductionAudioReader) BatchFactoryAudioRecord(ctx context.Context, owner, batchID, bookID, audioID string) (agentworkspace.BatchFactoryAudioRecord, bool, error) {
	owner, batchID = strings.TrimSpace(owner), strings.TrimSpace(batchID)
	bookID, audioID = strings.TrimSpace(bookID), strings.TrimSpace(audioID)
	if r == nil || r.batches == nil || r.files == nil || owner == "" || batchID == "" || bookID == "" || audioID == "" {
		return agentworkspace.BatchFactoryAudioRecord{}, false, nil
	}
	batch, err := r.batches.GetBatch(ctx, owner, batchID)
	if errors.Is(err, batchfactoryv11.ErrNotFound) {
		return agentworkspace.BatchFactoryAudioRecord{}, false, nil
	}
	if err != nil {
		return agentworkspace.BatchFactoryAudioRecord{}, false, err
	}
	for _, book := range batch.Books {
		if book.ID != bookID {
			continue
		}
		effective := batchfactoryv11.ResolveSettings(batch.SettingsState.Patch, book.SettingsState.Patch)
		if batchFactoryAudioSettingString(effective["ttsAudioArtifactId"]) != audioID {
			return agentworkspace.BatchFactoryAudioRecord{}, false, nil
		}
		duration := batchFactoryAudioSettingNumber(effective["audioDurationSeconds"])
		if duration <= 0 {
			return agentworkspace.BatchFactoryAudioRecord{}, false, nil
		}
		file, openErr := r.files.Open(audioID + ".mp3")
		if errors.Is(openErr, os.ErrNotExist) || errors.Is(openErr, localartifact.ErrInvalidID) {
			return agentworkspace.BatchFactoryAudioRecord{}, false, nil
		}
		if openErr != nil {
			return agentworkspace.BatchFactoryAudioRecord{}, false, openErr
		}
		_ = file.Close()
		return agentworkspace.BatchFactoryAudioRecord{
			ID: audioID, VersionID: audioID + ":" + strconv.FormatInt(book.SettingsState.Revision, 10),
			Status: "succeeded", DurationSeconds: duration,
		}, true, nil
	}
	return agentworkspace.BatchFactoryAudioRecord{}, false, nil
}

func batchFactoryAudioSettingString(raw json.RawMessage) string {
	var value string
	if len(raw) == 0 || json.Unmarshal(raw, &value) != nil {
		return ""
	}
	return strings.TrimSpace(value)
}

func batchFactoryAudioSettingNumber(raw json.RawMessage) float64 {
	var value float64
	if len(raw) == 0 || json.Unmarshal(raw, &value) != nil {
		return 0
	}
	return value
}
