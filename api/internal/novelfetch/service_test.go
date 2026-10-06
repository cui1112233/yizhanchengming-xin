package novelfetch

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type memoryStore struct {
	mu        sync.Mutex
	batches   map[string]Batch
	books     map[string]map[string]Book
	runs      map[string]Run
	records   map[string][]Record
	config    Config
	knowledge []KnowledgeEntry
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		batches: map[string]Batch{}, books: map[string]map[string]Book{},
		runs: map[string]Run{}, records: map[string][]Record{},
		config: Config{
			TextModelID: "text-model-1", MaxText: 12,
			TargetVersions: []string{"original", "ai1", "ai2"},
			RewriteProfiles: map[string]string{"ai1": "fast", "ai2": "story"},
			SensitiveReplacements: map[string]string{"敏感": "合规"},
			ChapterRemovePrefixes: []string{"第"},
			TrimLines: true, DropBlankLines: true,
		},
	}
}
func (m *memoryStore) CreateBatch(_ context.Context, value Batch) (Batch,error){m.mu.Lock();defer m.mu.Unlock();m.batches[value.ID]=value;if m.books[value.ID]==nil{m.books[value.ID]=map[string]Book{}};return value,nil}
func (m *memoryStore) GetBatch(_ context.Context,id string)(Batch,error){m.mu.Lock();defer m.mu.Unlock();v,ok:=m.batches[id];if !ok{return Batch{},ErrNotFound};return v,nil}
func (m *memoryStore) PutBook(_ context.Context,v Book)error{m.mu.Lock();defer m.mu.Unlock();if m.books[v.BatchID]==nil{m.books[v.BatchID]=map[string]Book{}};m.books[v.BatchID][v.Key]=v;return nil}
func (m *memoryStore) UpdateBook(ctx context.Context,v Book)error{return m.PutBook(ctx,v)}
func (m *memoryStore) GetBook(_ context.Context,batchID,key string)(Book,error){m.mu.Lock();defer m.mu.Unlock();v,ok:=m.books[batchID][key];if !ok{return Book{},ErrNotFound};return v,nil}
func (m *memoryStore) ListBooks(_ context.Context,batchID string)([]Book,error){m.mu.Lock();defer m.mu.Unlock();batch,ok:=m.batches[batchID];if !ok{return nil,ErrNotFound};out:=make([]Book,0,len(batch.BookKeys));for _,key:=range batch.BookKeys{out=append(out,m.books[batchID][key])};return out,nil}
func (m *memoryStore) CreateRun(_ context.Context,v Run)(Run,error){m.mu.Lock();defer m.mu.Unlock();m.runs[v.ID]=v;return v,nil}
func (m *memoryStore) UpdateRun(_ context.Context,v Run)error{m.mu.Lock();defer m.mu.Unlock();m.runs[v.ID]=v;return nil}
func (m *memoryStore) GetRun(_ context.Context,id string)(Run,error){m.mu.Lock();defer m.mu.Unlock();v,ok:=m.runs[id];if !ok{return Run{},ErrNotFound};return v,nil}
func (m *memoryStore) AppendRecord(_ context.Context,v Record)error{m.mu.Lock();defer m.mu.Unlock();m.records[v.RunID]=append(m.records[v.RunID],v);return nil}
func (m *memoryStore) ListRecords(_ context.Context,id string)([]Record,error){m.mu.Lock();defer m.mu.Unlock();return append([]Record(nil),m.records[id]...),nil}
func (m *memoryStore) GetConfig(context.Context)(Config,error){m.mu.Lock();defer m.mu.Unlock();return m.config,nil}
func (m *memoryStore) SaveConfig(_ context.Context,v Config)(Config,error){m.mu.Lock();defer m.mu.Unlock();m.config=v;return v,nil}
func (m *memoryStore) ListKnowledge(_ context.Context,kind string)([]KnowledgeEntry,error){m.mu.Lock();defer m.mu.Unlock();out:=[]KnowledgeEntry{};for _,v:=range m.knowledge{if kind==""||v.Kind==kind{out=append(out,v)}};return out,nil}
func (m *memoryStore) UpsertKnowledge(_ context.Context,v KnowledgeEntry)(KnowledgeEntry,error){m.mu.Lock();defer m.mu.Unlock();for i:=range m.knowledge{if m.knowledge[i].ID==v.ID{m.knowledge[i]=v;return v,nil}};m.knowledge=append(m.knowledge,v);return v,nil}
func (m *memoryStore) DeleteKnowledge(_ context.Context,kind,id string)error{m.mu.Lock();defer m.mu.Unlock();for i,v:=range m.knowledge{if v.ID==id&&(kind==""||v.Kind==kind){m.knowledge=append(m.knowledge[:i],m.knowledge[i+1:]...);return nil}};return ErrNotFound}

type fakeFetcher struct{calls int; result FetchResult; err error}
func (f *fakeFetcher) Fetch(context.Context,FetchRequest)(FetchResult,error){f.calls++;return f.result,f.err}

type flakyModel struct{mu sync.Mutex; calls map[string]int; failVersion string}
func (m *flakyModel) Rewrite(_ context.Context,req RewriteRequest)(string,error){m.mu.Lock();defer m.mu.Unlock();if m.calls==nil{m.calls=map[string]int{}};m.calls[req.Version]++;if req.Version==m.failVersion&&m.calls[req.Version]==1{return "",errors.New("temporary model failure")};return req.Version+"::"+req.Profile+"::"+req.Book.ProcessedText,nil}

type fakeDispatcher struct{items []RunDispatch; err error}
func (d *fakeDispatcher) EnqueueRun(_ context.Context,item RunDispatch)error{d.items=append(d.items,item);return d.err}

type fakeHandoff struct{calls int}
func (f *fakeHandoff) CreateFromNovelFetch(_ context.Context,request HandoffRequest)(HandoffResult,error){f.calls++;if len(request.Books)==0{return HandoffResult{},errors.New("books required")};return HandoffResult{BatchProjectID:99},nil}

type fakePublisher struct{calls int; last SubmitIntentRequest}
func (f *fakePublisher) CreateIntent(_ context.Context,request SubmitIntentRequest)(SubmitIntentResult,error){f.calls++;f.last=request;return SubmitIntentResult{IntentID:77,Status:"pending"},nil}

func fixedNow() time.Time { return time.Date(2026,10,7,1,0,0,0,time.FixedZone("CST",8*3600)) }

func TestCreateBatchSupportsMultiplePlatformsAndDeduplicatesWithinPlatform(t *testing.T){
	store:=newMemoryStore();service:=NewService(Options{Store:store,Now:fixedNow})
	batch,books,err:=service.CreateBatch(context.Background(),CreateBatchInput{Name:"multi",Groups:[]GroupInput{
		{Source:"阳光",PlatformID:"4",Books:[]BookInput{{BookID:"1001"},{BookID:"1001"},{BookID:"1002"}}},
		{Source:"常读",PlatformID:"2",Books:[]BookInput{{BookID:"1001"}}},
	}})
	if err!=nil{t.Fatal(err)}
	if len(books)!=3{t.Fatalf("books=%d want 3",len(books))}
	if len(batch.BookKeys)!=3{t.Fatalf("keys=%v",batch.BookKeys)}
	if books[0].Key==books[2].Key{t.Fatalf("cross-platform same book id must coexist")}
}

func TestFullOriginalIsPreservedWhileProcessingUsesMaxTextAndRules(t *testing.T){
	store:=newMemoryStore()
	fetcher:=&fakeFetcher{result:FetchResult{OriginalRaw:"第一章\n敏感内容ABCDEF\n后文保留XYZ",Title:"真实书名",Category:"男频",Genre:"都市"}}
	model:=&flakyModel{calls:map[string]int{}}
	dispatcher:=&fakeDispatcher{}
	service:=NewService(Options{Store:store,Fetcher:fetcher,Model:model,Dispatcher:dispatcher,Now:fixedNow})
	batch,_,err:=service.CreateBatch(context.Background(),CreateBatchInput{Groups:[]GroupInput{{Source:"阳光",PlatformID:"4",Books:[]BookInput{{BookID:"1001"}}}}});if err!=nil{t.Fatal(err)}
	run,err:=service.StartRun(context.Background(),batch.ID,time.Time{});if err!=nil{t.Fatal(err)}
	run,err=service.ExecuteRun(context.Background(),run.ID);if err!=nil{t.Fatal(err)}
	if run.Status!=StatusSucceeded{t.Fatalf("run=%s",run.Status)}
	book,err:=store.GetBook(context.Background(),batch.ID,"4_1001");if err!=nil{t.Fatal(err)}
	if book.OriginalRaw!="第一章\n敏感内容ABCDEF\n后文保留XYZ"{t.Fatalf("raw changed: %q",book.OriginalRaw)}
	if book.OriginalChars<=book.ProcessedChars{t.Fatalf("expected processed truncation: raw=%d processed=%d",book.OriginalChars,book.ProcessedChars)}
	if book.ProcessedText=="book.OriginalRaw"{t.Fatal("processed text must be independent")}
	if book.Versions["original"]!=book.ProcessedText{t.Fatalf("original version must use processed text")}
	if fetcher.calls!=1{t.Fatalf("fetch calls=%d",fetcher.calls)}
}

func TestPartialRewriteRetryDoesNotRefetchSuccessfulOriginalOrVersion(t *testing.T){
	store:=newMemoryStore()
	fetcher:=&fakeFetcher{result:FetchResult{OriginalRaw:"正文足够长1234567890"}}
	model:=&flakyModel{calls:map[string]int{},failVersion:"ai2"}
	dispatcher:=&fakeDispatcher{}
	service:=NewService(Options{Store:store,Fetcher:fetcher,Model:model,Dispatcher:dispatcher,Now:fixedNow})
	batch,_,_:=service.CreateBatch(context.Background(),CreateBatchInput{Groups:[]GroupInput{{Source:"阳光",PlatformID:"4",Books:[]BookInput{{BookID:"1"}}}}})
	run,err:=service.StartRun(context.Background(),batch.ID,time.Time{});if err!=nil{t.Fatal(err)}
	run,err=service.ExecuteRun(context.Background(),run.ID);if err!=nil{t.Fatal(err)}
	book,_:=store.GetBook(context.Background(),batch.ID,"4_1")
	if book.Status!=StatusPartialFailed{t.Fatalf("book status=%s",book.Status)}
	if model.calls["ai1"]!=1||model.calls["ai2"]!=1{t.Fatalf("calls=%v",model.calls)}
	book,err=service.RetryBook(context.Background(),run.ID,"4_1");if err!=nil{t.Fatal(err)}
	if book.Status!=StatusSucceeded{t.Fatalf("book status=%s error=%s",book.Status,book.Error)}
	if fetcher.calls!=1{t.Fatalf("fetch repeated: %d",fetcher.calls)}
	if model.calls["ai1"]!=1{t.Fatalf("successful ai1 repeated: %d",model.calls["ai1"])}
	if model.calls["ai2"]!=2{t.Fatalf("failed ai2 not retried: %d",model.calls["ai2"])}
	records,_:=service.Records(context.Background(),run.ID)
	attempts:=0
	for _,record:=range records{if record.Stage=="rewrite:ai2"{attempts++}}
	if attempts!=2{t.Fatalf("ai2 records=%d",attempts)}
}

func TestScheduledRunOnlyEnqueuesAndRejectsEarlyExecution(t *testing.T){
	store:=newMemoryStore();dispatcher:=&fakeDispatcher{}
	service:=NewService(Options{Store:store,Fetcher:&fakeFetcher{},Dispatcher:dispatcher,Now:fixedNow})
	batch,_,_:=service.CreateBatch(context.Background(),CreateBatchInput{Groups:[]GroupInput{{Source:"常读",PlatformID:"2",Books:[]BookInput{{BookID:"9"}}}}})
	when:=fixedNow().Add(time.Hour)
	run,err:=service.StartRun(context.Background(),batch.ID,when);if err!=nil{t.Fatal(err)}
	if run.Status!=StatusScheduled{t.Fatalf("status=%s",run.Status)}
	if len(dispatcher.items)!=1||!dispatcher.items[0].AvailableAt.Equal(when){t.Fatalf("dispatch=%+v",dispatcher.items)}
	if _,err:=service.ExecuteRun(context.Background(),run.ID);!errors.Is(err,ErrNotDue){t.Fatalf("err=%v",err)}
}

func TestNetworkSubmitCreatesIntentOnlyAndHandoffUsesBoundary(t *testing.T){
	store:=newMemoryStore();publisher:=&fakePublisher{};handoff:=&fakeHandoff{}
	service:=NewService(Options{Store:store,Fetcher:&fakeFetcher{result:FetchResult{OriginalRaw:"正文123456789012345"}},Model:&flakyModel{calls:map[string]int{}},Dispatcher:&fakeDispatcher{},Publisher:publisher,BatchFactory:handoff,Now:fixedNow})
	batch,_,_:=service.CreateBatch(context.Background(),CreateBatchInput{Groups:[]GroupInput{{Source:"阳光",PlatformID:"4",Books:[]BookInput{{BookID:"10"}}}}})
	run,err:=service.StartRun(context.Background(),batch.ID,time.Time{});if err!=nil{t.Fatal(err)}
	if _,err=service.ExecuteRun(context.Background(),run.ID);err!=nil{t.Fatal(err)}
	intent,err:=service.RequestNetworkSubmit(context.Background(),SubmitIntentRequest{RunID:run.ID,BookKey:"4_10",Version:"ai1",PublishingAccountID:5});if err!=nil{t.Fatal(err)}
	if intent.Status!="pending"||publisher.calls!=1{t.Fatalf("intent=%+v calls=%d",intent,publisher.calls)}
	if publisher.last.BatchID!=batch.ID{t.Fatalf("batch boundary lost: %+v",publisher.last)}
	result,err:=service.HandoffToBatchFactory(context.Background(),run.ID);if err!=nil{t.Fatal(err)}
	if result.BatchProjectID!=99||handoff.calls!=1{t.Fatalf("handoff=%+v calls=%d",result,handoff.calls)}
}
