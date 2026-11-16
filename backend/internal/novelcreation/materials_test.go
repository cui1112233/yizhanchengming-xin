package novelcreation

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

type fakeDocumentObjects struct {
	objects map[string][]byte
	failPut bool
}

func (f *fakeDocumentObjects) Put(_ context.Context, o DocumentObject, r io.Reader) error {
	if f.failPut {
		return errors.New("offline")
	}
	b, _ := io.ReadAll(r)
	f.objects[o.Key] = b
	return nil
}
func (f *fakeDocumentObjects) Open(_ context.Context, o DocumentObject) (io.ReadCloser, error) {
	b, ok := f.objects[o.Key]
	if !ok {
		return nil, errors.New("missing")
	}
	return io.NopCloser(bytes.NewReader(b)), nil
}
func (f *fakeDocumentObjects) Delete(_ context.Context, o DocumentObject) error {
	delete(f.objects, o.Key)
	return nil
}
func materialService(t *testing.T) (*Service, *MemoryStore, Project) {
	t.Helper()
	st := NewMemoryStore()
	s := NewService(st, NewMemoryRuntime(), nil)
	p, e := s.CreateProject(context.Background(), "alice", CreateProjectInput{Title: "Book", Genre: "幻想", ModelID: "fake"})
	if e != nil {
		t.Fatal(e)
	}
	return s, st, p
}
func TestMaterialDurableParseAndSelection(t *testing.T) {
	ctx := context.Background()
	s, _, p := materialService(t)
	if _, e := s.UploadDocument(ctx, "alice", p.ID, "a.txt", strings.NewReader("hello")); e == nil {
		t.Fatal("missing TOS accepted")
	}
	objects := &fakeDocumentObjects{objects: map[string][]byte{}}
	s.ConfigureDocuments(objects, nil)
	j, e := s.UploadDocument(ctx, "alice", p.ID, "a.txt", strings.NewReader("第一章 初见\n你好。\n第二章 重逢\n再见。"))
	if e != nil || j.Status != "queued" {
		t.Fatalf("%+v %v", j, e)
	}
	if _, e = s.GetDocument(ctx, "bob", p.ID, j.ID); e == nil {
		t.Fatal("cross account")
	}
	j, e = s.ParseDocument(ctx, "alice", p.ID, j.ID)
	if e != nil || j.Status != "ready" || len(j.Chapters) != 2 {
		t.Fatalf("%+v %v", j, e)
	}
	selected, e := s.SelectDocument(ctx, "alice", p.ID, j.ID, DocumentSelection{ExpectedVersion: j.Version, ChapterIDs: []string{j.Chapters[1].ID, j.Chapters[0].ID}})
	if e != nil || len(selected.SelectedChapterIDs) != 2 || !strings.HasPrefix(selected.Content, "第二章") {
		t.Fatalf("%+v %v", selected, e)
	}
	if _, e = s.SelectDocument(ctx, "alice", p.ID, j.ID, DocumentSelection{ExpectedVersion: 99, ChapterIDs: []string{j.Chapters[0].ID}}); e == nil {
		t.Fatal("stale source accepted")
	}
	bad, e := s.UploadDocument(ctx, "alice", p.ID, "bad.pdf", strings.NewReader("broken"))
	if e != nil {
		t.Fatal(e)
	}
	bad, e = s.ParseDocument(ctx, "alice", p.ID, bad.ID)
	if e == nil || bad.Status != "failed" || len(bad.Chapters) > 0 || objects.objects[bad.Object.Key] == nil {
		t.Fatalf("lost failed source %+v %v", bad, e)
	}
}
