package videogen

import (
	"context"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/providerconfig"
)

type providerRecordSourceFake struct { resolved providerconfig.Resolved; owner,kind,provider string }
func (f *providerRecordSourceFake) Resolve(_ context.Context,owner,kind,provider string)(providerconfig.Resolved,error){ f.owner,f.kind,f.provider=owner,kind,provider; return f.resolved,nil }

func TestConfigSourceMapsEncryptedProviderRecord(t *testing.T){
	source:=&providerRecordSourceFake{resolved:providerconfig.Resolved{Record:providerconfig.Record{Provider:ProviderPersonalAPI,Model:"yd2.0-mini",CreateURL:"https://create",TasksURL:"https://tasks",ResultURL:"https://result/{id}",Enabled:true},APIKey:"secret"}}
	adapter:=ConfigStoreSource{Store:source}
	cfg,err:=adapter.Resolve(context.Background(),"owner",ProviderPersonalAPI)
	if err!=nil{t.Fatal(err)}
	if source.owner!="owner"||source.kind!=providerconfig.KindVideo||source.provider!=ProviderPersonalAPI{t.Fatalf("source=%#v",source)}
	if cfg.APIKey!="secret"||cfg.Model!="yd2.0-mini"||cfg.ResultURL!="https://result/{id}"{t.Fatalf("cfg=%#v",cfg)}
}

func TestConfigSourceRejectsDisabledConfig(t *testing.T){
	source:=&providerRecordSourceFake{resolved:providerconfig.Resolved{Record:providerconfig.Record{Provider:ProviderPersonalAPI,Enabled:false},APIKey:"secret"}}
	if _,err:=(ConfigStoreSource{Store:source}).Resolve(context.Background(),"owner",ProviderPersonalAPI);err==nil{t.Fatal("expected disabled config error")}
}
