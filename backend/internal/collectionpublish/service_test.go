package collectionpublish

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"qiantie/backend/internal/n8comic"
)

type testRecords map[string]Record

func (s testRecords) Load(_ context.Context, r PublishRequest) (Record, bool, error) {
	v, ok := s[string(r.Kind)+r.CollectionID]
	return v, ok, nil
}
func (s testRecords) Save(_ context.Context, r PublishRequest, v Record) error {
	s[string(r.Kind)+r.CollectionID] = v
	return nil
}

func TestCollectionPublishKindOrderAndIdempotency(t *testing.T) {
	for _, k := range []CollectionKind{CollectionKindComic, CollectionKindText} {
		t.Run(string(k), func(t *testing.T) {
			r := PublishRequest{Kind: k, Owner: "a", BatchID: "batch", BookID: "book", CollectionID: "c", Producer: "a", OrganizationID: "org", CategoryID: "category", AudioCategoryID: 5, AudioID: 7, WorkIDs: []int64{31}}
			calls := 0
			svc := Service{Records: testRecords{}, Submit: func(_ context.Context, in n8comic.PublishRequest) (n8comic.PublishResult, error) {
				calls++
				if string(in.Images[0].Content) != "1" || string(in.Images[1].Content) != "2" {
					t.Fatal("wrong image order")
				}
				return n8comic.PublishResult{CarouselID: int64(calls), ImageIDs: []int64{6, 7}}, nil
			}}
			prepare := func(_ context.Context, _ *PublishRequest) (Collection, error) {
				return Collection{Kind: k, ID: "c", Name: "作品", Pages: []Page{{Ordinal: 2, Image: n8comic.ImageInput{Content: []byte("2")}}, {Ordinal: 1, Image: n8comic.ImageInput{Content: []byte("1")}}}}, nil
			}
			first, err := svc.Publish(context.Background(), r, prepare)
			if err != nil || !first.Confirmed() {
				t.Fatalf("first=%+v %v", first, err)
			}
			same, err := svc.Publish(context.Background(), r, prepare)
			if err != nil || same.IntentID != first.IntentID || calls != 1 {
				t.Fatalf("repeat=%+v calls=%d %v", same, calls, err)
			}
			r.Force = true
			r.RequestKey = "second"
			second, err := svc.Publish(context.Background(), r, prepare)
			if err != nil || second.IntentID == first.IntentID || calls != 2 {
				t.Fatalf("force=%+v %v", second, err)
			}
			_, err = svc.Publish(context.Background(), r, prepare)
			if err != nil || calls != 2 {
				t.Fatalf("replayed force=%d %v", calls, err)
			}
			r.Collections = []Reference{{Kind: CollectionKindComic, ID: "c"}, {Kind: CollectionKindText, ID: "t"}}
			_, err = svc.Publish(context.Background(), r, prepare)
			if !errors.Is(err, ErrMixedKinds) || calls != 2 {
				t.Fatalf("mixed calls=%d %v", calls, err)
			}
		})
	}
}

func TestCollectionPublishPendingAndKindMismatch(t *testing.T) {
	calls := 0
	svc := Service{Records: testRecords{}, Submit: func(context.Context, n8comic.PublishRequest) (n8comic.PublishResult, error) {
		calls++
		return n8comic.PublishResult{}, nil
	}}
	r := PublishRequest{Kind: CollectionKindText, Owner: "owner", BatchID: "batch", BookID: "b", CollectionID: "c", Producer: "p", OrganizationID: "o", CategoryID: "g", AudioCategoryID: 1, AudioID: 2, WorkIDs: []int64{3}}
	coll := Collection{Kind: CollectionKindComic, ID: "c", Name: "作品", Pages: []Page{{Ordinal: 1}, {Ordinal: 2}}}
	prepare := func(context.Context, *PublishRequest) (Collection, error) { return coll, nil }
	if _, err := svc.Publish(context.Background(), r, prepare); !errors.Is(err, ErrMixedKinds) || calls != 0 {
		t.Fatal(err, calls)
	}
	coll.Kind = CollectionKindText
	for range 2 {
		rec, err := svc.Publish(context.Background(), r, prepare)
		if err != nil || rec.Status != "pending" || rec.Confirmed() || calls != 1 {
			t.Fatalf("pending=%+v %v calls=%d", rec, err, calls)
		}
	}
}

func TestCollectionPublishAcceptedWithoutCarouselIDStaysPending(t *testing.T) {
	for _, k := range []CollectionKind{CollectionKindComic, CollectionKindText} {
		t.Run(string(k), func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.URL.Path != "/material/open/carousel/create" {
					t.Errorf("unexpected path %s", r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, `{"code":"SUCCESS","data":{}}`)
			}))
			defer server.Close()
			client := n8comic.NewClient(n8comic.Config{BaseURL: server.URL, APIKey: "test", HTTPClient: server.Client()})
			records := testRecords{}
			svc := Service{Records: records, Submit: func(ctx context.Context, in n8comic.PublishRequest) (n8comic.PublishResult, error) {
				ids := []int64{7, 8}
				id, err := client.CreateCarousel(ctx, n8comic.CreateCarouselRequest{ProducerAdminName: in.ProducerAdminName, Name: in.Name, WorkIDs: in.WorkIDs, AudioCategoryID: in.AudioCategoryID, AudioID: in.AudioID, ImageIDs: ids})
				return n8comic.PublishResult{CarouselID: id, ImageIDs: ids}, err
			}}
			req := PublishRequest{Kind: k, Owner: "o", BatchID: "b", BookID: "s", CollectionID: "c", Producer: "a", OrganizationID: "org", CategoryID: "tag", AudioCategoryID: 3, AudioID: 77, WorkIDs: []int64{12}, RunID: "run-first", Attempt: 1}
			prepare := func(context.Context, *PublishRequest) (Collection, error) {
				return Collection{Kind: k, ID: "c", Name: "测试", Pages: []Page{{Ordinal: 1}, {Ordinal: 2}}}, nil
			}
			for i := range 2 {
				rec, err := svc.Publish(context.Background(), req, prepare)
				if err != nil || rec.Confirmed() || rec.Status != "pending" || len(rec.ImageIDs) != 2 || rec.RunID != "run-first" || rec.Attempt != 1 || calls != 1 {
					t.Fatalf("attempt %d calls=%d rec=%+v err=%v", i, calls, rec, err)
				}
				req.RunID = "later"
				req.Attempt = 2
			}
		})
	}
}

func TestCollectionPublishConfirmedReuseAfterMappingChangeBindsRun(t *testing.T) {
	ctx := context.Background()
	records := testRecords{}
	rec := Record{Status: "published", Kind: CollectionKindText, BookID: "b", CollectionID: "c", IntentID: "intent", RequestKey: "manual", MappingRevision: "old", CarouselID: 9, PageCount: 2, ImageIDs: []int64{3, 4}}
	r := PublishRequest{Kind: CollectionKindText, Owner: "owner", BatchID: "batch", BookID: "b", CollectionID: "c", MappingRevision: "changed", RunID: "run", Attempt: 2, RequestKey: "auto"}
	records.Save(ctx, r, rec)
	svc := Service{Records: records}
	got, err := svc.Publish(ctx, r, func(context.Context, *PublishRequest) (Collection, error) {
		t.Fatal("unexpected new upload")
		return Collection{}, nil
	})
	if err != nil || !got.Confirmed() || got.IntentID != rec.IntentID || got.RunID != "run" || got.RequestKey != "auto" || got.Attempt != 2 || got.LastPublished == nil || got.LastPublished.RequestKey != "manual" {
		t.Fatalf("derived=%+v err=%v", got, err)
	}
	stored, _, _ := records.Load(ctx, r)
	if stored.RunID != got.RunID {
		t.Fatalf("not persisted: %+v", stored)
	}
}

func TestCollectionPublishPendingReadbackResumesWithoutSending(t *testing.T) {
	ctx := context.Background()
	records := testRecords{}
	r := PublishRequest{Kind: CollectionKindText, Owner: "owner", BatchID: "batch", BookID: "b", CollectionID: "c", RunID: "new-run", Attempt: 2}
	records.Save(ctx, r, Record{Status: "pending", Kind: r.Kind, BookID: "b", CollectionID: "c", IntentID: "pending-intent", ImageIDs: []int64{3, 4}, PageCount: 2, LastPublished: &Record{Status: "published", CarouselID: 9, ImageIDs: []int64{3, 4}}})
	reads := 0
	svc := Service{Records: records, Readback: func(ctx context.Context, in n8comic.ReceiptLookup) (n8comic.PublishResult, error) {
		reads++
		id := int64(9)
		if reads > 1 {
			id = 17
		}
		return n8comic.PublishResult{CarouselID: id, ImageIDs: []int64{3, 4}}, nil
	}}
	pending, ok, err := svc.Status(ctx, r)
	if err != nil || !ok || pending.Status != "pending" || pending.LastPublished.CarouselID != 9 {
		t.Fatal(pending, err)
	}
	got, _, err := svc.Status(ctx, r)
	if err != nil || !got.Confirmed() || got.CarouselID != 17 || got.IntentID != "pending-intent" || reads != 2 {
		t.Fatal(got, err)
	}
	got, _, err = svc.Status(ctx, r)
	if err != nil || !got.Confirmed() || reads != 2 {
		t.Fatal(got, err)
	}
}

func TestCollectionPublishFailedRepublishNeedsForceWhenHistoryConfirmed(t *testing.T) {
	ctx := context.Background()
	r := PublishRequest{Kind: CollectionKindText, Owner: "owner", BatchID: "batch", BookID: "b", CollectionID: "c"}
	prior := Record{Kind: r.Kind, BookID: r.BookID, CollectionID: r.CollectionID, IntentID: "accepted-first", Status: "published", CarouselID: 11, ImageIDs: []int64{1, 2}}
	records := testRecords{}
	records.Save(ctx, r, Record{Kind: r.Kind, BookID: r.BookID, CollectionID: r.CollectionID, Status: "upload_incomplete", IntentID: "failed-forced", LastPublished: &prior})
	svc := Service{Records: records}
	result, err := svc.Publish(ctx, r, func(context.Context, *PublishRequest) (Collection, error) {
		t.Fatal("retried a republish without confirmation")
		return Collection{}, nil
	})
	if err != nil || !result.Confirmed() || result.IntentID != "accepted-first" {
		t.Fatal(result, err)
	}
	current, _, _ := svc.Status(ctx, r)
	if current.IntentID != "failed-forced" {
		t.Fatal("ordinary history lookup overwrote latest", current)
	}
}
