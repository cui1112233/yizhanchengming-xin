package httpapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/pipeline"
)

const maxRequestBody = 1 << 20

type handler struct {
	deps Dependencies
}

type createIntakeRequest struct {
	Name   string             `json:"name"`
	Groups []createBookGroup  `json:"groups"`
}

type createBookGroup struct {
	Source     string            `json:"source"`
	PlatformID string            `json:"platformId"`
	Books      []createBookInput `json:"books"`
}

type createBookInput struct {
	BookID string `json:"bookId"`
	Title  string `json:"title"`
	Gender string `json:"gender"`
	Style  string `json:"style"`
}

type executeRequest struct {
	MaxText int `json:"maxText"`
}

type batchProjectRequest struct {
	Name  string `json:"name"`
	RunAt string `json:"runAt"`
}

type intakeResponse struct {
	ID        int64         `json:"id"`
	Name      string        `json:"name"`
	Status    intake.Status `json:"status"`
	CreatedAt time.Time     `json:"createdAt,omitempty"`
	UpdatedAt time.Time     `json:"updatedAt,omitempty"`
}

type bookResponse struct {
	ID             int64             `json:"id"`
	IntakeID       int64             `json:"intakeId"`
	Source         string            `json:"source"`
	PlatformID     string            `json:"platformId"`
	ExternalBookID string            `json:"bookId"`
	Title          string            `json:"title"`
	BodyRef        string            `json:"bodyRef,omitempty"`
	Category       string            `json:"category,omitempty"`
	Genre          string            `json:"genre,omitempty"`
	Gender         string            `json:"gender,omitempty"`
	GenderSource   string            `json:"genderSource,omitempty"`
	Style          string            `json:"style,omitempty"`
	Status         intake.BookStatus `json:"status"`
	ErrorMessage   string            `json:"errorMessage,omitempty"`
	UpdatedAt      time.Time         `json:"updatedAt,omitempty"`
}

type projectResponse struct {
	ID       int64  `json:"id"`
	IntakeID int64  `json:"intakeId"`
	Name     string `json:"name"`
}

type runResponse struct {
	ID             int64            `json:"id"`
	BatchProjectID int64            `json:"batchProjectId"`
	RunAt          time.Time        `json:"runAt"`
	Status         intake.RunStatus `json:"status"`
}

func (h handler) createIntake(w http.ResponseWriter, r *http.Request) {
	if h.deps.Intakes == nil {
		writeError(w, http.StatusServiceUnavailable, "intake service unavailable")
		return
	}
	var request createIntakeRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	input := intake.CreateIntakeInput{Name: request.Name, Groups: make([]intake.BookGroup, 0, len(request.Groups))}
	for _, group := range request.Groups {
		converted := intake.BookGroup{Source: group.Source, PlatformID: group.PlatformID, Books: make([]intake.BookInput, 0, len(group.Books))}
		for _, book := range group.Books {
			converted.Books = append(converted.Books, intake.BookInput{BookID: book.BookID, Title: book.Title, Gender: book.Gender, Style: book.Style})
		}
		input.Groups = append(input.Groups, converted)
	}
	created, books, err := h.deps.Intakes.CreateIntake(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"intake": toIntakeResponse(created), "books": toBookResponses(books)})
}

func (h handler) listIntakes(w http.ResponseWriter, r *http.Request) {
	if h.deps.Reader == nil {
		writeError(w, http.StatusServiceUnavailable, "intake reader unavailable")
		return
	}
	rows, err := h.deps.Reader.ListIntakes(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "读取 intake 列表失败")
		return
	}
	result := make([]intakeResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, toIntakeResponse(row))
	}
	writeJSON(w, http.StatusOK, map[string]any{"intakes": result})
}

func (h handler) executeIntake(w http.ResponseWriter, r *http.Request) {
	if h.deps.Intakes == nil {
		writeError(w, http.StatusServiceUnavailable, "intake service unavailable")
		return
	}
	id, err := parsePositiveID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	request := executeRequest{MaxText: 4000}
	if r.ContentLength != 0 {
		if err := decodeJSON(w, r, &request); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	if request.MaxText < 100 || request.MaxText > 100000 {
		writeError(w, http.StatusBadRequest, "maxText 必须在 100–100000 之间")
		return
	}
	result, err := h.deps.Intakes.ExecuteIntake(r.Context(), id, request.MaxText)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h handler) listBooks(w http.ResponseWriter, r *http.Request) {
	if h.deps.Reader == nil {
		writeError(w, http.StatusServiceUnavailable, "intake reader unavailable")
		return
	}
	id, err := parsePositiveID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	rows, err := h.deps.Reader.ListBooks(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "读取书籍列表失败")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"books": toBookResponses(rows)})
}

func (h handler) createBatchProject(w http.ResponseWriter, r *http.Request) {
	if h.deps.Pipeline == nil {
		writeError(w, http.StatusServiceUnavailable, "pipeline service unavailable")
		return
	}
	id, err := parsePositiveID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	var request batchProjectRequest
	if r.ContentLength != 0 {
		if err := decodeJSON(w, r, &request); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	create := pipeline.CreateRequest{IntakeID: id, Name: request.Name}
	if raw := strings.TrimSpace(request.RunAt); raw != "" {
		parsed, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "runAt 必须是 RFC3339 时间")
			return
		}
		create.RunAt = parsed
	}
	result, err := h.deps.Pipeline.Create(r.Context(), create)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"project": projectResponse{ID: result.Project.ID, IntakeID: result.Project.IntakeID, Name: result.Project.Name},
		"run": runResponse{ID: result.Run.ID, BatchProjectID: result.Run.BatchProjectID, RunAt: result.Run.RunAt, Status: result.Run.Status},
	})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBody)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("请求 JSON 无效: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return fmt.Errorf("请求体只能包含一个 JSON 对象")
		}
		return fmt.Errorf("请求 JSON 无效: %w", err)
	}
	return nil
}

func parsePositiveID(raw string) (int64, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("id 必须是正整数")
	}
	return id, nil
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func toIntakeResponse(value intake.Intake) intakeResponse {
	return intakeResponse{ID: value.ID, Name: value.Name, Status: value.Status, CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt}
}

func toBookResponses(rows []intake.Book) []bookResponse {
	result := make([]bookResponse, 0, len(rows))
	for _, value := range rows {
		result = append(result, bookResponse{
			ID: value.ID, IntakeID: value.IntakeID, Source: value.Source, PlatformID: value.PlatformID,
			ExternalBookID: value.ExternalBookID, Title: value.Title, BodyRef: value.BodyRef,
			Category: value.Category, Genre: value.Genre, Gender: value.Gender, GenderSource: value.GenderSource,
			Style: value.Style, Status: value.Status, ErrorMessage: value.ErrorMessage, UpdatedAt: value.UpdatedAt,
		})
	}
	return result
}
