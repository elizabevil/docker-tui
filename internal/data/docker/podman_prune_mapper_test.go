package docker

import (
	"encoding/json"
	"testing"
)

func TestPodmanReportErrorPreservesPartialFailure(t *testing.T) {
	if err := podmanReportError(nil); err != nil {
		t.Fatalf("nil report error = %v", err)
	}
	err := podmanReportError(json.RawMessage(`"volume is in use"`))
	if err == nil || err.Error() != "volume is in use" {
		t.Fatalf("report error = %v", err)
	}
}
