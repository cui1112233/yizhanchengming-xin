package batchfactory

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/novel"
)

const defaultMaxTxt = 4000

type IntakeBookInput struct {
	BookID       string
	ManualGender novel.Gender
	Style        string
}

type IntakeGroupInput struct {
	PlatformID   string
	PlatformName string
	MaxTxt       int
	Books        []IntakeBookInput
}

type CreateIntakeInput struct {
	Owner  string
	Title  string
	Groups []IntakeGroupInput
}

type IntakeRecord struct {
	ID    string
	Owner string
	Title string
	Books []IntakeBookRecord
}

type IntakeBookRecord struct {
	PlatformID   string
	PlatformName string
	MaxTxt       int
	BookID       string
	ManualGender novel.Gender
	Style        string
}

type CreateIntakeResult struct {
	IntakeID   string `json:"intake_id"`
	GroupCount int    `json:"group_count"`
	BookCount  int    `json:"book_count"`
}

type IntakeStore interface {
	CreateIntake(context.Context, IntakeRecord) error
}

type IntakeService struct {
	Store IntakeStore
	NewID func() string
}

func (s IntakeService) Create(ctx context.Context, input CreateIntakeInput) (CreateIntakeResult, error) {
	if s.Store == nil {
		return CreateIntakeResult{}, errors.New("intake store is required")
	}
	owner := strings.TrimSpace(input.Owner)
	if owner == "" {
		return CreateIntakeResult{}, errors.New("owner is required")
	}
	if len(input.Groups) == 0 {
		return CreateIntakeResult{}, errors.New("at least one bookstore group is required")
	}

	record := IntakeRecord{Owner: owner, Title: strings.TrimSpace(input.Title)}
	seen := make(map[string]struct{})
	groupCount := 0
	for groupIndex, group := range input.Groups {
		platformID := strings.TrimSpace(group.PlatformID)
		platformName := strings.TrimSpace(group.PlatformName)
		if platformID == "" {
			return CreateIntakeResult{}, fmt.Errorf("group %d platform_id is required", groupIndex+1)
		}
		if platformName == "" {
			return CreateIntakeResult{}, fmt.Errorf("group %d platform_name is required", groupIndex+1)
		}
		if len(group.Books) == 0 {
			return CreateIntakeResult{}, fmt.Errorf("group %d must contain at least one book", groupIndex+1)
		}
		maxTxt := group.MaxTxt
		if maxTxt <= 0 {
			maxTxt = defaultMaxTxt
		}
		groupCount++
		for bookIndex, book := range group.Books {
			bookID := strings.TrimSpace(book.BookID)
			if bookID == "" {
				return CreateIntakeResult{}, fmt.Errorf("group %d book %d book_id is required", groupIndex+1, bookIndex+1)
			}
			if book.ManualGender != novel.GenderUnknown && book.ManualGender != novel.GenderMale && book.ManualGender != novel.GenderFemale {
				return CreateIntakeResult{}, fmt.Errorf("group %d book %d manual_gender is invalid", groupIndex+1, bookIndex+1)
			}
			key := platformID + "\x00" + bookID
			if _, exists := seen[key]; exists {
				return CreateIntakeResult{}, fmt.Errorf("duplicate book_id %q in platform %q", bookID, platformID)
			}
			seen[key] = struct{}{}
			record.Books = append(record.Books, IntakeBookRecord{
				PlatformID: platformID, PlatformName: platformName, MaxTxt: maxTxt,
				BookID: bookID, ManualGender: book.ManualGender, Style: strings.TrimSpace(book.Style),
			})
		}
	}
	if len(record.Books) == 0 {
		return CreateIntakeResult{}, errors.New("at least one book is required")
	}

	if s.NewID != nil {
		record.ID = strings.TrimSpace(s.NewID())
	} else {
		record.ID = newIntakeID()
	}
	if record.ID == "" {
		return CreateIntakeResult{}, errors.New("intake id is required")
	}
	if err := s.Store.CreateIntake(ctx, record); err != nil {
		return CreateIntakeResult{}, err
	}
	return CreateIntakeResult{IntakeID: record.ID, GroupCount: groupCount, BookCount: len(record.Books)}, nil
}

func newIntakeID() string {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return ""
	}
	return hex.EncodeToString(value[:])
}
