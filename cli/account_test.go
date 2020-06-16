package cli

import "testing"

func TestAccount(t *testing.T) {
	cli := NewCLI()

	cli.TestCommand("account new")
	cli.TestCommand("account new -w /tmp/walletPath")

	cli.TestCommand("account new -n 10 -w /tmp/walletPath")

	cli.TestCommand("account list")
	cli.TestCommand("account list -w /tmp/empty")

}
