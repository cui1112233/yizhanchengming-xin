package videogen

import (
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/media"
)

type factoryFake struct{ provider Provider; resolved string; requestModel string }
func (f *factoryFake) ForModel(_ context.Context, owner, model, fallback string)(Provider,string,error){ f.requestModel=model; return f.provider,f.resolved,nil }

type providerFake struct{ submits []Request; polls int; submit Task; poll Task }
func (p *providerFake) Submit(_ context.Context, req Request)(Task,error){ p.submits=append(p.submits,req); return p.submit,nil }
func (p *providerFake) Poll(_ context.Context, task Task)(Task,error){ p.polls++; return p.poll,nil }

type referenceFake struct{ urls map[string]string; ids []string }
func (r *referenceFake) Resolve(_ context.Context, owner,id string)(media.ResolvedAsset,error){ r.ids=append(r.ids,id); return media.ResolvedAsset{Asset:media.Reference{ID:id,MediaType:media.TypeImage},PreviewURL:r.urls[id]},nil }

type downloadFake struct{ url string; data string; contentType string }
func (d *downloadFake) Download(_ context.Context, raw string)(DownloadedMedia,error){ d.url=raw; return DownloadedMedia{Body:io.NopCloser(strings.NewReader(d.data)),ContentType:d.contentType,ContentLength:int64(len(d.data))},nil }

type ingestFake struct{ input media.IngestInput; asset media.Asset }
func (i *ingestFake) Ingest(_ context.Context,input media.IngestInput)(media.Asset,error){ i.input=input; if i.asset.ID==""{i.asset=media.Asset{ID:"asset_video",MediaType:media.TypeVideo}}; return i.asset,nil }

func TestServiceGeneratesVideoFromMediaAssetReferencesAndIngestsTOS(t *testing.T){
	provider:=&providerFake{submit:Task{ID:"provider-task",State:StateQueued},poll:Task{ID:"provider-task",State:StateSucceeded,MediaURL:"https://cdn.example.com/result.mp4",DurationMs:10000}}
	factory:=&factoryFake{provider:provider,resolved:ProviderPersonalAPI}
	refs:=&referenceFake{urls:map[string]string{"asset_img_1":"https://signed.example/ref1.png"}}
	downloader:=&downloadFake{data:"video-bytes",contentType:"video/mp4"}
	ingestor:=&ingestFake{}
	service:=Service{Providers:factory,References:refs,Downloader:downloader,Ingestor:ingestor,PollInterval:time.Millisecond,Sleep:func(context.Context,time.Duration)error{return nil}}
	asset,err:=service.Generate(context.Background(),"owner",GenerateInput{ToolCallID:"tool_1",Model:"yd2-mini-video",Prompt:"镜头推进",Duration:10,AspectRatio:"9:16",Resolution:"720p",ReferenceMediaAssetIDs:[]string{"asset_img_1"}})
	if err!=nil{t.Fatal(err)}
	if asset.ID!="asset_video"{t.Fatalf("asset=%#v",asset)}
	if len(provider.submits)!=1||provider.submits[0].ReferenceURLs[0]!="https://signed.example/ref1.png"{t.Fatalf("submit=%#v",provider.submits)}
	if provider.polls!=1{t.Fatalf("polls=%d",provider.polls)}
	if downloader.url!="https://cdn.example.com/result.mp4"{t.Fatalf("download=%q",downloader.url)}
	if ingestor.input.Owner!="owner"||ingestor.input.MediaType!=media.TypeVideo||ingestor.input.ContentType!="video/mp4"{t.Fatalf("ingest=%#v",ingestor.input)}
	var metadata map[string]any
	if err:=json.Unmarshal(ingestor.input.Metadata,&metadata);err!=nil{t.Fatal(err)}
	if metadata["tool_call_id"]!="tool_1"||metadata["provider"]!=ProviderPersonalAPI||metadata["model"]!="yd2-mini-video"{t.Fatalf("metadata=%#v",metadata)}
}

func TestServiceAcceptsProviderImmediateSuccess(t *testing.T){
	provider:=&providerFake{submit:Task{State:StateSucceeded,MediaURL:"https://cdn.example.com/direct.mp4"}}
	service:=Service{Providers:&factoryFake{provider:provider,resolved:ProviderYFAISeedance},References:&referenceFake{urls:map[string]string{}},Downloader:&downloadFake{data:"v",contentType:"video/mp4"},Ingestor:&ingestFake{}}
	if _,err:=service.Generate(context.Background(),"owner",GenerateInput{ToolCallID:"tool",Model:"seedance-2-0-official",Prompt:"x",Duration:5});err!=nil{t.Fatal(err)}
	if provider.polls!=0{t.Fatalf("polls=%d",provider.polls)}
}

func TestServiceRejectsNonImageReferenceAsset(t *testing.T){
	refs:=referenceFunc(func(context.Context,string,string)(media.ResolvedAsset,error){return media.ResolvedAsset{Asset:media.Reference{ID:"asset_video",MediaType:media.TypeVideo},PreviewURL:"https://signed.example/v.mp4"},nil})
	service:=Service{Providers:&factoryFake{provider:&providerFake{}},References:refs,Downloader:&downloadFake{},Ingestor:&ingestFake{}}
	if _,err:=service.Generate(context.Background(),"owner",GenerateInput{ToolCallID:"tool",Model:"yd2.0-mini",Prompt:"x",Duration:5,ReferenceMediaAssetIDs:[]string{"asset_video"}});err==nil{t.Fatal("expected image reference validation")}
}

type referenceFunc func(context.Context,string,string)(media.ResolvedAsset,error)
func (f referenceFunc) Resolve(ctx context.Context,owner,id string)(media.ResolvedAsset,error){return f(ctx,owner,id)}

func TestServiceStopsWhenProviderFails(t *testing.T){
	provider:=&providerFake{submit:Task{ID:"x",State:StateQueued},poll:Task{ID:"x",State:StateFailed}}
	service:=Service{Providers:&factoryFake{provider:provider,resolved:ProviderPersonalAPI},References:&referenceFake{urls:map[string]string{}},Downloader:&downloadFake{},Ingestor:&ingestFake{},Sleep:func(context.Context,time.Duration)error{return nil}}
	if _,err:=service.Generate(context.Background(),"owner",GenerateInput{ToolCallID:"tool",Model:"yd2.0-mini",Prompt:"x",Duration:5});err==nil{t.Fatal("expected provider failure")}
}
