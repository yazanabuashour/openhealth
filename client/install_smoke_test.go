package client_test

import (
	"os/exec"
	"testing"
)

func TestClientExampleBuilds(t *testing.T) {
	t.Parallel()

	cmd := exec.Command("go", "build", "-o", t.TempDir()+"/client-summary", "./examples/client_summary")
	cmd.Dir = ".."
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build example: %v\n%s", err, output)
	}
}
