package cli

import "testing"

func TestPay(t *testing.T) {
	cli := NewCLI()

	cli.TestCommand("pay 1")
	cli.TestCommand("pay 2 -u Gwei")
	cli.TestCommand("pay 2 -c ./../config.toml -w ../")
	cli.TestCommand("pay 10 --from 0xDB2C9C06E186D58EFe19f213b3d5FaF8B8c99481 --to 0x7cbdfE7371f56A8f996d9EBa7c66AEddB3f221f3 -i http://192.168.168.33:8501 -w ../")

	// no keystore file
	cli.TestCommand("pay 10 --from 0x0000000000000000000000000000000000000001 --to 0x7cbdfE7371f56A8f996d9EBa7c66AEddB3f221f3 -i http://192.168.168.33:8501 -w ../")

	// walletPath is a file
	cli.TestCommand("pay 10 -c ./../config.toml -w test.txt")
}
