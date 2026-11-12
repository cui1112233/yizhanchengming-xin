package aicut

import (
	"bytes"
	"context"
	"encoding/binary"
	"testing"
)

func TestSilencePCMUsesChannelEnergyAndPadding(t *testing.T) {
	var b bytes.Buffer
	for ms := 0; ms < 3000; ms++ {
		var v int16
		if ms < 1000 || ms >= 2000 {
			v = 12000
		}
		binary.Write(&b, binary.LittleEndian, []int16{v, -v})
	}
	got, err := AnalyzeSilencePCM(context.Background(), &b, 1000, 2, SilenceOptions{MinSilenceMS: 400, ThresholdDB: -40, PaddingMS: 100})
	if err != nil || got.AllSilent || len(got.Intervals) != 1 || got.Intervals[0] != (SilenceInterval{1100, 1900}) {
		t.Fatalf("%+v %v", got, err)
	}
}
func TestSilencePCMBoundsCancellationAndAllSilent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := AnalyzeSilencePCM(ctx, bytes.NewReader(make([]byte, 2000)), 1000, 1, SilenceOptions{}); err == nil {
		t.Fatal("cancel accepted")
	}
	got, err := AnalyzeSilencePCM(context.Background(), bytes.NewReader(make([]byte, 2000)), 1000, 1, SilenceOptions{})
	if err != nil || !got.AllSilent {
		t.Fatalf("%+v %v", got, err)
	}
	for _, opts := range []SilenceOptions{{MinSilenceMS: -1}, {ThresholdDB: 1}, {PaddingMS: 2001}} {
		if _, err := NormalizeSilenceOptions(opts); err == nil {
			t.Fatal("invalid bounds accepted")
		}
	}
}
