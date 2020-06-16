package cli

import "testing"

func TestBalance(t *testing.T) {
	cli := NewCLI()

	cli.TestCommand("balance --help")

	cli.TestCommand("balance 0xDB2C9C06E186D58EFe19f213b3d5FaF8B8c99481 -c ./../config.toml")

	cli.TestCommand("balance 0x01 002 003 0x004 0xDB2C9C06E186D58EFe19f213b3d5FaF8B8c99481 -c ./../config.toml")

}
