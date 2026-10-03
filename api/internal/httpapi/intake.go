package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/batchfactory"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/novel"
)

type IntakeCreator interface {
	Create(context.Context, batchfactory.CreateIntakeInput) (batchfactory.CreateIntakeResult, error)
}

type OwnerResolver func(*http.Request) (string, error)

type intakeHandler struct {
	creator IntakeCreator
	owner   OwnerResolver
}

type createIntakeRequest struct {
	Title  string               `json:"title"`
	Groups []intakeGroupRequest `json:"groups"`
}

type intakeGroupRequest struct {
	PlatformID   string              `json:"platform_id"`
	PlatformName string              `json:"platform_name"`
	MaxTxt       int                 `json:"max_txt"`
	Books        []intakeBookRequest `json:"books"`
}

type intakeBookRequest struct {
	BookID       string `json:"book_id"`
	ManualGender string `json:"manual_gender"`
	Style        string `json:"style"`
}

func NewIntakeHandler(creator IntakeCreator, owner OwnerResolver) http.Handler {
	return &intakeHandler{creator: creator, owner: owner}
}

func (h *intakeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if h.creator == nil || h.owner == nil {
		http.Error(w, "intake service unavailable", http.StatusServiceUnavailable)
		return
	}
	owner, err := h.owner(r)
	if err != nil || strings.TrimSpace(owner) == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var request createIntakeRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	groups, err := convertIntakeGroups(request.Groups)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	input := batchfactory.CreateIntakeInput{
		Owner: strings.TrimSpace(owner),
		Title: request.Title,
		Groups: groups,
	}
	result, err := h.creator.Create(r.Context(), input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(result)
}

func convertIntakeGroups(groups []intakeGroupRequest) ([]batchfactory.IntakeGroupInput, error) {
	convertedGroups := make([]batchfactory.IntakeGroupInput, 0, len(groups))
	for _, group := range groups {
		converted := batchfactory.IntakeGroupInput{
			PlatformID: group.PlatformID,
			PlatformName: group.PlatformName,
			MaxTxt: group.MaxTxt,
		}
		for _, book := range group.Books {
			gender, err := parseManualGender(book.ManualGender)
			if err != nil {
				return nil, err
			}
			converted.Books = append(converted.Books, batchfactory.IntakeBookInput{
				BookID: book.BookID,
				ManualGender: gender,
				Style: book.Style,
			})
		}
		convertedGroups = append(convertedGroups, converted)
	}
	return convertedGroups, nil
}

func parseManualGender(value string) (novel.Gender, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "":
		return novel.GenderUnknown, nil
	case "male", "男", "男频", "男生", "男性", "男向":
		return novel.GenderMale, nil
	case "female", "女", "女频", "女生", "女性", "女向":
		return novel.GenderFemale, nil
	default:
		return novel.GenderUnknown, errors.New("manual_gender must be male/female, 男频/女频, or empty")
	}
}
