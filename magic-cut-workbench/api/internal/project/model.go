package project

import "time"

type Project struct {
	ID             string    `json:"id"`
	Owner          string    `json:"-"`
	Name           string    `json:"name"`
	Width          int       `json:"width"`
	Height         int       `json:"height"`
	FrameRate      int       `json:"frameRate"`
	IdempotencyKey string    `json:"-"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type CreateInput struct {
	Name           string `json:"name"`
	Width          int    `json:"width"`
	Height         int    `json:"height"`
	FrameRate      int    `json:"frameRate"`
	IdempotencyKey string `json:"idempotencyKey"`
}
