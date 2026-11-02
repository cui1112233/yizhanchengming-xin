package httpapi

import (
	"errors"
	"net/http"

	"qiantie/backend/internal/batchfactoryv11"
)

// textProviderInput remains a wire-compatible, ignored field for older clients.
// Provider selection and credentials are resolved once by Go from the account
// configuration; request bodies must never be able to override them.
type textProviderInput struct {
	Endpoint string `json:"endpoint"`
	APIKey   string `json:"apiKey"`
	Model    string `json:"model"`
}

func directorForRequest(service *batchfactoryv11.DirectorService, input *textProviderInput) (*batchfactoryv11.DirectorService, error) {
	_ = input
	if service == nil || service.Store == nil {
		return nil, batchfactoryv11.ErrUnavailable
	}
	return service, nil
}

func registerDirectorRoutes(mux *http.ServeMux, service *batchfactoryv11.DirectorService, store batchfactoryv11.Store) {
	mux.HandleFunc("GET /api/batch-factory/v11/batches/{batchId}/books/{bookId}/director-prompt-candidate", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		candidate, err := service.DirectorPromptCandidate(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"candidate": candidate})
	})
	mux.HandleFunc("POST /api/batch-factory/v11/batches/{batchId}/books/{bookId}/director-prompt-candidate", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input struct {
			ExpectedRevision int64              `json:"expectedRevision"`
			TextProvider     *textProviderInput `json:"textProvider"`
		}
		if !decodeJSON(w, r, &input) {
			return
		}
		runner, err := directorForRequest(service, input.TextProvider)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		candidate, err := runner.CreateDirectorPromptCandidate(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"), input.ExpectedRevision)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"candidate": candidate})
	})
	mux.HandleFunc("POST /api/batch-factory/v11/batches/{batchId}/books/{bookId}/videos/{videoId}/director-prompt-candidate/accept", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input struct {
			CandidateID      string `json:"candidateId"`
			ExpectedRevision int64  `json:"expectedRevision"`
		}
		if !decodeJSON(w, r, &input) {
			return
		}
		video, err := service.AcceptDirectorPromptCandidate(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"), r.PathValue("videoId"), input.CandidateID, input.ExpectedRevision)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"video": video})
	})
	mux.HandleFunc("DELETE /api/batch-factory/v11/batches/{batchId}/books/{bookId}/director-prompt-candidate", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input struct {
			ExpectedRevision int64 `json:"expectedRevision"`
		}
		if !decodeJSON(w, r, &input) {
			return
		}
		if err := service.CancelDirectorPromptCandidate(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"), input.ExpectedRevision); err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"cancelled": true})
	})
	mux.HandleFunc("POST /api/batch-factory/v11/batches/{batchId}/books/{bookId}/working-front/viral", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input struct {
			Content        string                         `json:"content"`
			DerivedOpening batchfactoryv11.PresetSnapshot `json:"derivedOpening"`
			TextProvider   *textProviderInput             `json:"textProvider"`
		}
		if !decodeJSON(w, r, &input) {
			return
		}
		runner, err := directorForRequest(service, input.TextProvider)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		candidate, err := runner.RewriteWorkingFront(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"), input.Content, input.DerivedOpening)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"candidate": candidate})
	})
	mux.HandleFunc("POST /api/batch-factory/v11/batches/{batchId}/books/{bookId}/hook", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input struct {
			TextProvider *textProviderInput `json:"textProvider"`
		}
		if !decodeJSON(w, r, &input) {
			return
		}
		runner, err := directorForRequest(service, input.TextProvider)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		value, err := runner.RunHook(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"hook": value})
	})
	mux.HandleFunc("POST /api/batch-factory/v11/batches/{batchId}/books/{bookId}/opening-variants", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input struct {
			OpeningMeta  batchfactoryv11.PresetSnapshot `json:"openingMeta"`
			TextProvider *textProviderInput             `json:"textProvider"`
		}
		if !decodeJSON(w, r, &input) {
			return
		}
		runner, err := directorForRequest(service, input.TextProvider)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		variants, err := runner.RunOpeningVariants(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"), input.OpeningMeta)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"variants": variants})
	})
	mux.HandleFunc("POST /api/batch-factory/v11/batches/{batchId}/books/{bookId}/hooks/{hookId}/approve", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		value, err := service.ApproveHook(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"), r.PathValue("hookId"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"hook": value})
	})
	mux.HandleFunc("POST /api/batch-factory/v11/batches/{batchId}/books/{bookId}/director", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input struct {
			TextProvider      *textProviderInput `json:"textProvider"`
			SmartUnifiedStyle string             `json:"smartUnifiedStyle"`
		}
		if !decodeJSON(w, r, &input) {
			return
		}
		runner, err := directorForRequest(service, input.TextProvider)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		value, err := runner.RunDirectorWithSmartUnifiedStyle(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"), input.SmartUnifiedStyle)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"directorRevision": value})
	})
	mux.HandleFunc("POST /api/batch-factory/v11/batches/{batchId}/director", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input struct {
			TextProvider *textProviderInput `json:"textProvider"`
		}
		if !decodeJSON(w, r, &input) {
			return
		}
		runner, err := directorForRequest(service, input.TextProvider)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		batchID := r.PathValue("batchId")
		batch, err := store.GetBatch(r.Context(), owner, batchID)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		type failure struct {
			BookID string `json:"bookId"`
			Error  string `json:"error"`
		}
		revisions := []batchfactoryv11.DirectorRevision{}
		failures := []failure{}
		for _, book := range batch.Books {
			revision, runErr := runner.RunDirector(r.Context(), owner, batchID, book.ID)
			if runErr != nil {
				failures = append(failures, failure{BookID: book.ID, Error: safeDirectorError(runErr)})
				continue
			}
			revisions = append(revisions, revision)
		}
		status := http.StatusCreated
		if len(revisions) == 0 && len(failures) > 0 {
			status = http.StatusConflict
		}
		writeJSON(w, status, map[string]any{"directorRevisions": revisions, "failures": failures})
	})
}

func safeDirectorError(err error) string {
	switch {
	case errors.Is(err, batchfactoryv11.ErrNotFound):
		return "not found"
	case errors.Is(err, batchfactoryv11.ErrConflict):
		return "director prerequisites not satisfied"
	case errors.Is(err, batchfactoryv11.ErrInvalid):
		return "invalid director input or output"
	case errors.Is(err, batchfactoryv11.ErrUnavailable):
		return "director unavailable"
	default:
		return "director failed"
	}
}
