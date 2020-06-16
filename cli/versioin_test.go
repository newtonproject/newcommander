package cli

import "testing"

func TestVersion(t *testing.T) {
	cli := NewCLI()

	cli.TestCommand("version")
}
