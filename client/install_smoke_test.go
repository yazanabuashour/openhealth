package client_test

import (
	"os"
	"os/exec"
	"testing"

	"github.com/yazanabuashour/openhealth/client"
)

func TestClientExampleRunsAgainstLocalRuntime(t *testing.T) {
	t.Parallel()

	cmd := exec.Command("go", "run", "./examples/client_summary")
	cmd.Dir = ".."
	cmd.Env = append(os.Environ(), client.EnvDataDir+"="+t.TempDir())
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run example: %v\n%s", err, output)
	}
}
