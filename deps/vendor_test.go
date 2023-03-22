package deps

import "os"
import "os/exec"
import "testing"

func TestVendored(t *testing.T) {
	os.Chdir("vendor_test")

	run := func(c string, args ...string) {
		cmd := exec.Command(c, args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			os.Stdout.Write(output)
			t.Fatal(err)
		}
	}

	run("go", "mod", "tidy")
	run("go", "mod", "vendor")
	run("go", "test")
	run("git", "checkout", "go.mod")
}
