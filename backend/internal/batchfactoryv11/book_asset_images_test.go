package batchfactoryv11

import "testing"

func TestValidBookAssetImageInputAcceptsDoubaoLocalExecutorArtifact(t *testing.T) {
	input := CreateBookAssetImageInput{
		StorageRef: "lea_example.png",
		MediaType:  "image/png",
		Source:     "doubao_local_executor",
	}

	got, err := validBookAssetImageInput(input)
	if err != nil {
		t.Fatalf("validBookAssetImageInput() error = %v", err)
	}
	if got.StorageRef != "lea_example.png" || got.MediaType != "image/png" || got.Source != "doubao_local_executor" {
		t.Fatalf("validBookAssetImageInput() = %+v", got)
	}
}
