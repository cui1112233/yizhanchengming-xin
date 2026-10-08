package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/workshop"
)

func (h handler) getWorkshop(w http.ResponseWriter, r *http.Request) {
	if h.deps.Workshop == nil {
		writeError(w, http.StatusServiceUnavailable, "workshop service unavailable")
		return
	}
	id, err := parsePositiveID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	value, err := h.deps.Workshop.Snapshot(r.Context(), id)
	if err != nil {
		workshopError(w, err, "读取 Workshop 失败")
		return
	}
	books := make([]map[string]any, 0, len(value.Books))
	for _, book := range value.Books {
		books = append(books, map[string]any{"id": book.ID, "intakeId": book.IntakeID, "source": book.Source, "platformId": book.PlatformID, "bookId": book.ExternalBookID, "title": book.Title, "category": book.Category, "genre": book.Genre, "gender": book.Gender, "style": book.Style, "status": book.Status, "errorMessage": book.ErrorMessage, "originalText": book.OriginalText})
	}
	writeJSON(w, http.StatusOK, map[string]any{"intake": toIntakeResponse(value.Intake), "books": books, "settings": json.RawMessage(value.Settings), "prompts": value.Prompts})
}

func (h handler) saveWorkshop(w http.ResponseWriter, r *http.Request) {
	if h.deps.Workshop == nil {
		writeError(w, http.StatusServiceUnavailable, "workshop service unavailable")
		return
	}
	id, err := parsePositiveID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	var body struct {
		Settings json.RawMessage `json:"settings"`
	}
	if err := decodeJSON(w, r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	settings, err := h.deps.Workshop.Save(r.Context(), id, body.Settings)
	if err != nil {
		workshopError(w, err, "保存 Workshop 配置失败")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"intakeId": id, "settings": json.RawMessage(settings)})
}

func workshopError(w http.ResponseWriter, err error, message string) {
	if errors.Is(err, workshop.ErrNotFound) {
		writeError(w, http.StatusNotFound, "任务不存在")
		return
	}
	writeError(w, http.StatusUnprocessableEntity, message)
}

func (h handler) restoreWorkshopBook(w http.ResponseWriter, r *http.Request) {
	if h.deps.Intakes == nil {
		writeError(w, http.StatusServiceUnavailable, "intake service unavailable")
		return
	}
	intakeID, err := parsePositiveID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	bookID, err := parsePositiveID(r.PathValue("bookId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	var body struct {
		MaxText int `json:"maxText"`
	}
	if r.ContentLength != 0 {
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	if body.MaxText == 0 {
		body.MaxText = 4000
	}
	if body.MaxText < 100 || body.MaxText > 100000 {
		writeError(w, http.StatusBadRequest, "maxText 必须在 100–100000 之间")
		return
	}
	book, err := h.deps.Intakes.RestoreBook(r.Context(), intakeID, bookID, body.MaxText)
	if errors.Is(err, intake.ErrBatchProjectArchived) {
		writeJSON(w, http.StatusConflict, map[string]any{"code": "BATCH_PROJECT_ARCHIVED", "message": "项目已归档，请先恢复后再恢复原文"})
		return
	}
	if err != nil {
		h.writeServiceError(w, r, http.StatusUnprocessableEntity, "BOOK_RESTORE_FAILED", "原文恢复失败", "intake", "restore_book", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"book": map[string]any{"id": book.ID, "intakeId": book.IntakeID, "bookId": book.ExternalBookID, "title": book.Title, "status": book.Status, "originalText": book.OriginalText, "errorMessage": book.ErrorMessage}})
}
