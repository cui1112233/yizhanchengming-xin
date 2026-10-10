package httpapi

import (
	"archive/zip"
	"bytes"
	"context"
	"net/http/httptest"
	"testing"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
)

func TestWriteShuihuoProjectExportContainsSourceAndProjectData(t *testing.T) {
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "导出检查", Books: []batchfactoryv11.CreateBookInput{{Title: "原文", SourceText: "第一句。"}}})
	if err != nil {
		t.Fatal(err)
	}
	var response = httptest.NewRecorder()
	if err := writeShuihuoProjectExport(response, httptest.NewRequest("GET", "/", nil), localartifact.NewStore(t.TempDir(), 1<<20), batch); err != nil {
		t.Fatal(err)
	}
	if response.Code != 200 || response.Header().Get("Content-Type") != "application/zip" {
		t.Fatalf("status=%d content-type=%q", response.Code, response.Header().Get("Content-Type"))
	}
	archive, err := zip.NewReader(bytes.NewReader(response.Body.Bytes()), int64(response.Body.Len()))
	if err != nil {
		t.Fatal(err)
	}
	entries := map[string]bool{}
	for _, entry := range archive.File {
		entries[entry.Name] = true
	}
	for _, name := range []string{"source.txt", "segments.json", "project.json"} {
		if !entries[name] {
			t.Fatalf("export missing %s", name)
		}
	}
}
