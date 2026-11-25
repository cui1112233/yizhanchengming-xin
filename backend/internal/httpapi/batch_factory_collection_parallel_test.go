package httpapi

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	cp "qiantie/backend/internal/collectionpublish"
	"qiantie/backend/internal/n8comic"
)

type blockingCollectionClient struct {
	started chan struct{}
	release chan struct{}
	calls   atomic.Int32
}

func (*blockingCollectionClient) Configured() bool { return true }
func (c *blockingCollectionClient) Publish(ctx context.Context, _ n8comic.PublishRequest) (n8comic.PublishResult, error) {
	id := c.calls.Add(1)
	c.started <- struct{}{}
	select {
	case <-c.release:
		return n8comic.PublishResult{CarouselID: int64(id), ImageIDs: []int64{71, 72}}, nil
	case <-ctx.Done():
		return n8comic.PublishResult{}, ctx.Err()
	}
}

func TestPublisherAllowsIndependentCollectionsAndDeduplicatesSameCollection(t *testing.T) {
	for _, same := range []bool{false, true} {
		t.Run(fmt.Sprintf("same_collection_%t", same), func(t *testing.T) {
			ctx := context.Background()
			store := batchfactoryv11.NewMemoryStore()
			batch, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{Title: "parallel", Books: []batchfactoryv11.CreateBookInput{{Title: "one"}, {Title: "two"}}})
			if err != nil {
				t.Fatal(err)
			}
			client := &blockingCollectionClient{started: make(chan struct{}, 2), release: make(chan struct{})}
			p := &BatchFactoryN8ComicPublisher{Store: store, Client: client}
			request := func(i int) cp.PublishRequest {
				return cp.PublishRequest{Kind: cp.CollectionKindText, Owner: "alice", BatchID: batch.ID, BookID: batch.Books[i].ID, CollectionID: fmt.Sprintf("collection-%d", i), Producer: "maker", OrganizationID: "1", AudioCategoryID: 1, AudioID: 2, WorkIDs: []int64{3}}
			}
			first, second := request(0), request(1)
			if same {
				second = first
			}
			publish := func(r cp.PublishRequest) error {
				rec, err := p.collectionService().Publish(ctx, r, func(_ context.Context, in *cp.PublishRequest) (cp.Collection, error) {
					return cp.Collection{Kind: in.Kind, ID: in.CollectionID, Name: in.CollectionID, Pages: []cp.Page{{Ordinal: 1}, {Ordinal: 2}}}, nil
				})
				if err == nil && !rec.Confirmed() {
					return fmt.Errorf("unconfirmed receipt")
				}
				return err
			}
			done := make(chan error, 2)
			go func() { done <- publish(first) }()
			select {
			case <-client.started:
			case <-time.After(2 * time.Second):
				close(client.release)
				t.Fatal("first collection did not start")
			}
			go func() { done <- publish(second) }()
			independentStarted := false
			if !same {
				select {
				case <-client.started:
					independentStarted = true
				case <-time.After(2 * time.Second):
				}
			}
			close(client.release)
			for range 2 {
				if err := <-done; err != nil {
					t.Error(err)
				}
			}
			if !same && !independentStarted {
				t.Fatal("independent collection waited behind another book's network upload")
			}
			want := int32(2)
			if same {
				want = 1
			}
			if client.calls.Load() != want {
				t.Fatalf("submissions=%d, want %d", client.calls.Load(), want)
			}
		})
	}
}
