package aicut

import (
	"encoding/json"
	"math"
	"testing"
)

func TestAudioPanValidationAndPersistence(t *testing.T) {
	for _, pan := range []float64{-1, -.375, 0, .25, 1} {
		value := AudioProcessing{Pan: pan}
		if !validAudioProcessing(&value) {
			t.Fatalf("valid pan %v rejected", pan)
		}
		encoded, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		var restored AudioProcessing
		if err := json.Unmarshal(encoded, &restored); err != nil {
			t.Fatal(err)
		}
		if restored.Pan != pan {
			t.Fatalf("pan did not persist: %v", restored.Pan)
		}
	}
	for _, pan := range []float64{-1.01, 1.01, math.NaN(), math.Inf(1), math.Inf(-1)} {
		if validAudioProcessing(&AudioProcessing{Pan: pan}) {
			t.Fatalf("invalid pan %v accepted", pan)
		}
	}
}
