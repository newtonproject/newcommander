package cli

import "testing"

func TestSign(t *testing.T) {
	cli := NewCLI()

	cli.TestCommand("build")
	cli.TestCommand("sign tx")
	cli.TestCommand("submit tx.sign")
}
