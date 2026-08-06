package docker

import (
	"bytes"
	"context"
	"testing"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

func TestValidateImageTransferRequest(t *testing.T) {
	valid := []runtimeapi.ImageTransferRequest{
		{Operation: runtimeapi.ImageTransferTag, Source: "sha256:id", Destination: "example/app:v1"},
		{Operation: runtimeapi.ImageTransferPush, Source: "example/app:v1", Destination: "example/app:v1"},
		{Operation: runtimeapi.ImageTransferSave, Source: "sha256:id", Path: "/tmp/app.tar"},
		{Operation: runtimeapi.ImageTransferLoad, Path: "/tmp/app.tar"},
	}
	for _, request := range valid {
		if err := request.Validate(); err != nil {
			t.Fatalf("validate %#v: %v", request, err)
		}
	}
	if err := (runtimeapi.ImageTransferRequest{Operation: runtimeapi.ImageTransferSave}).Validate(); err == nil {
		t.Fatal("save without source and path was accepted")
	}
}

func TestProgressReaderReportsBytes(t *testing.T) {
	output := make(chan runtimeapi.ImageTransferEvent, 1)
	reader := &progressReader{ctx: context.Background(), output: output, status: "loading", total: 4, reader: bytes.NewBufferString("data")}
	data := make([]byte, 4)
	if _, err := reader.Read(data); err != nil {
		t.Fatal(err)
	}
	event := <-output
	if event.Progress == nil || event.Progress.Current != 4 || event.Progress.Total != 4 {
		t.Fatalf("progress event = %#v", event)
	}
}
