package cli

import "testing"

func TestInit(t *testing.T) {
	cli := NewCLI()

	cli.TestCommand("init")
}
