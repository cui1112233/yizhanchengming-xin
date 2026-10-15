// Command video-management-smoke verifies the Go connector against the real
// video-management target without using the legacy browser worker.
//
// Required environment variables:
//
//	QIANTIE_VIDEO_SMOKE_USERNAME
//	QIANTIE_VIDEO_SMOKE_PASSWORD
//
// Optional:
//
//	QIANTIE_VIDEO_CONNECTOR_TARGET_BASE_URL (defaults to http://two.121w.com)
//	QIANTIE_VIDEO_SMOKE_BOOK_ID (also performs a book readback)
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"qiantie/backend/internal/videomanagementconnector"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "go connector smoke failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	username := strings.TrimSpace(os.Getenv("QIANTIE_VIDEO_SMOKE_USERNAME"))
	password := os.Getenv("QIANTIE_VIDEO_SMOKE_PASSWORD")
	if username == "" || password == "" {
		return fmt.Errorf("QIANTIE_VIDEO_SMOKE_USERNAME and QIANTIE_VIDEO_SMOKE_PASSWORD are required")
	}
	service := &videomanagementconnector.TargetHTTPService{
		BaseURL: env("QIANTIE_VIDEO_CONNECTOR_TARGET_BASE_URL", "http://two.121w.com"),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	session, err := service.Login(ctx, videomanagementconnector.LoginInput{Username: username, Password: password})
	if err != nil {
		return err
	}
	if err := service.CheckSession(ctx, session); err != nil {
		return err
	}
	catalogs, err := service.RefreshCatalogs(ctx, session)
	if err != nil {
		return err
	}
	profiles, err := service.RefreshProfiles(ctx, session)
	if err != nil {
		return err
	}
	result := map[string]any{
		"runtime":         "go",
		"session":         "valid",
		"catalogs":        catalogs.Raw != nil,
		"profiles":        profiles != nil,
		"book_readback":   false,
		"upload_readback": false,
	}
	if bookID := strings.TrimSpace(os.Getenv("QIANTIE_VIDEO_SMOKE_BOOK_ID")); bookID != "" {
		book, err := service.GetBook(ctx, session, bookID, 1, 10)
		if err != nil {
			return err
		}
		result["book_readback"] = book != nil
		result["book_id"] = bookID
	}
	if uploadPath := strings.TrimSpace(os.Getenv("QIANTIE_VIDEO_SMOKE_UPLOAD_FILE")); uploadPath != "" {
		bookID := strings.TrimSpace(os.Getenv("QIANTIE_VIDEO_SMOKE_BOOK_ID"))
		if bookID == "" {
			return fmt.Errorf("QIANTIE_VIDEO_SMOKE_BOOK_ID is required when upload file is provided")
		}
		content, err := os.ReadFile(uploadPath)
		if err != nil {
			return fmt.Errorf("read smoke upload file: %w", err)
		}
		submitted, err := service.Submit(ctx, session, videomanagementconnector.SubmissionInput{
			BookID: bookID, Filename: filepath.Base(uploadPath), Content: content,
		})
		if err != nil {
			return err
		}
		readbackID := submitted.ID
		if readbackID == "" {
			readbackID = bookID
		}
		readback, err := service.GetSubmission(ctx, session, readbackID)
		if err != nil {
			return err
		}
		result["upload_readback"] = readback.ReadbackRaw != nil
		result["upload_id"] = readbackID
	}
	encoded, _ := json.Marshal(result)
	fmt.Println(string(encoded))
	return nil
}

func env(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}
