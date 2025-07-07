package inspector

import (
	"encoding/json"
	"os"
	"testing"
)

func TestGetProcInfo(t *testing.T) {
	info, err := GetProcInfo(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("ProcInfo: %s", b)
}
