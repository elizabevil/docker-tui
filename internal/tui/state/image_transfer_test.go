package state

import (
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

func TestImageTransferStateCancelsPreviousGeneration(t *testing.T) {
	var transfer ImageTransferState
	firstContext, firstGeneration := transfer.Begin(runtimeapi.ImageTransferRequest{Operation: runtimeapi.ImageTransferSave}, audit.Trace{})
	_, secondGeneration := transfer.Begin(runtimeapi.ImageTransferRequest{Operation: runtimeapi.ImageTransferLoad}, audit.Trace{})
	select {
	case <-firstContext.Done():
	default:
		t.Fatal("starting a transfer did not cancel the previous context")
	}
	if transfer.Current(firstGeneration) || !transfer.Current(secondGeneration) {
		t.Fatalf("transfer generations = %d, %d", firstGeneration, secondGeneration)
	}
}
