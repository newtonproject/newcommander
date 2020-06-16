package cli

import "testing"

func TestFaucet(t *testing.T) {
	cli := NewCLI()

	cli.TestCommand("faucet 0xDB2C9C06E186D58EFe19f213b3d5FaF8B8c99481 -i https://rpc1.newchain.newtonproject.org")

	cli.TestCommand("faucet 0xDB2C9C06E186D58EFe19f213b3d5FaF8B8c99481 -c ../config.toml")
}
