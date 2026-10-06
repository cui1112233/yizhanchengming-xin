package shuihuo

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrNotFound         = errors.New("shuihuo: not found")
	ErrForbidden        = errors.New("shuihuo: forbidden")
	ErrInvalid          = errors.New("shuihuo: invalid")
	ErrSmartUnavailable = errors.New("shuihuo: smart segmentation unavailable")
)

type SegmentationStatus string

const (
	SegmentationDraft     SegmentationStatus = "draft"
	SegmentationConfirmed SegmentationStatus = "confirmed"
)

type Actor struct {
	UserID          int64
	TeamID          int64
	BypassOwnership bool
}

type Project struct {
	ID                 int64              `json:"id"`
	OwnerUserID        int64              `json:"-"`
	TeamID             int64              `json:"-"`
	Name               string             `json:"name"`
	ProductionMode     string             `json:"productionMode"`
	SourceText         string             `json:"sourceText"`
	SegmentationStatus SegmentationStatus `json:"segmentationStatus"`
	CreatedAt          time.Time          `json:"createdAt"`
	UpdatedAt          time.Time          `json:"updatedAt"`
}

type Segment struct {
	ID           int64     `json:"id"`
	ProjectID    int64     `json:"projectId"`
	Position     int       `json:"position"`
	SourceText   string    `json:"sourceText"`
	SubtitleText string    `json:"subtitleText"`
	Speaker      string    `json:"speaker"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type Candidate struct {
	Text    string `json:"text"`
	Speaker string `json:"speaker,omitempty"`
}

type ReadModel struct {
	Project  Project   `json:"project"`
	Segments []Segment `json:"segments"`
}

type CreateProjectInput struct {
	Name           string `json:"name"`
	ProductionMode string `json:"productionMode,omitempty"`
}

type SegmentationInput struct {
	Text string `json:"text,omitempty"`
}
type FixedSegmentationInput struct {
	Text            string `json:"text,omitempty"`
	LinesPerSegment int    `json:"linesPerSegment"`
}
type SegmentInput struct {
	SourceText   string `json:"sourceText"`
	SubtitleText string `json:"subtitleText,omitempty"`
	Speaker      string `json:"speaker,omitempty"`
}

func CanAccess(actor Actor, project Project) bool {
	if actor.BypassOwnership {
		return true
	}
	if actor.UserID > 0 && actor.UserID == project.OwnerUserID {
		return true
	}
	return actor.TeamID > 0 && project.TeamID > 0 && actor.TeamID == project.TeamID
}

func normalizeSpeaker(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "旁白"
	}
	return value
}
