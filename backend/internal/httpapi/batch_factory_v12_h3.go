package httpapi

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"qiantie/backend/internal/batchfactoryv11"
)

func registerV12H3Routes(mux *http.ServeMux, service *batchfactoryv11.H3KernelService, director *batchfactoryv11.DirectorService) {
	mux.HandleFunc("POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/h3/director", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input struct {
			batchfactoryv11.H3DirectorRunRequest
			TextProvider *textProviderInput `json:"textProvider"`
		}
		if !decodeJSON(w, r, &input) {
			return
		}
		runner, err := directorForRequest(director, input.TextProvider)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		if runner == nil {
			writeStoreError(w, batchfactoryv11.ErrUnavailable)
			return
		}
		revision, err := runner.RunH3Director(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"), input.H3DirectorRunRequest)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"directorRevision": revision})
	})
	mux.HandleFunc("POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/h3/audio-measurement", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, (64<<20)*2)
		var input struct {
			AudioBase64        string                             `json:"audio_base64"`
			DirectorRevisionID string                             `json:"director_revision_id"`
			TTSFingerprint     string                             `json:"tts_fingerprint"`
			Lines              []batchfactoryv11.H3LineAudioInput `json:"lines"`
		}
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid H3 audio measurement request: " + err.Error()})
			return
		}
		if input.Lines != nil {
			if input.AudioBase64 != "" {
				writeStoreError(w, batchfactoryv11.ErrInvalid)
				return
			}
			measurement, err := service.MeasureLineAudio(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"), input.DirectorRevisionID, input.TTSFingerprint, input.Lines)
			if err != nil {
				writeStoreError(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, map[string]any{"audio_asset_id": measurement.Measurement.AssetID, "audio_measurement": measurement})
			return
		}
		audio, err := base64.StdEncoding.DecodeString(input.AudioBase64)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid H3 audio bytes"})
			return
		}
		measurement, err := service.MeasureAudio(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"), audio)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"audio_asset_id": measurement.Measurement.AssetID, "audio_measurement": measurement})
	})
	mux.HandleFunc("POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/h3/compile", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input batchfactoryv11.H3KernelCompileRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid H3 compile request: " + err.Error()})
			return
		}
		result, err := service.Compile(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"), input)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, result)
	})
}

// registerV12H3TraceRoute is kept separate because trace is a read-only
// query. It is safe to expose in slice 1 while H3 writes remain gated.
func registerV12H3TraceRoute(mux *http.ServeMux, service *batchfactoryv11.H3KernelService) {
	mux.HandleFunc("GET /api/batch-factory/v12/batches/{batchId}/books/{bookId}/h3/trace", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		trace, err := service.Trace(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"), r.URL.Query().Get("compilationId"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, trace)
	})
}
