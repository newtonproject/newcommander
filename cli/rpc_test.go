package cli

import "testing"

func TestRPC(t *testing.T) {
	cli := NewCLI()

	cli.TestCommand("rpc net_version")
	cli.TestCommand("rpc eth_getBlockByNumber latest true")
	cli.TestCommand("rpc eth_getBlockByNumber 1024 true")
	cli.TestCommand("rpc eth_getTransactionByHash 0xcad4299fd6516c7f66cbb5ae70114d9c06d73908a66c9165dbfc1e36fb67d892")
	cli.TestCommand("rpc eth_getTransactionReceipt 0xd9f1a3a4c54b6218c848e9a246faa570ea592a9f28d8b14e3ad4035398875de7")
	cli.TestCommand("rpc eth_getTransactionCount 0xc94770007dda54cF92009BFF0dE90c06F603a09f latest")
	cli.TestCommand("rpc eth_getTransactionCount 0xc94770007dda54cF92009BFF0dE90c06F603a09f 1024")
	cli.TestCommand("rpc eth_getBalance 0xc94770007dda54cF92009BFF0dE90c06F603a09f latest")
	cli.TestCommand("rpc eth_estimateGas {}")
	cli.TestCommand(`rpc eth_estimateGas '{"from":"0xdf9106238879143e914ad78d4ff4b4fa6b3b1648","gasPrice":"0x64","to":"0xdf9106238879143e914ad78d4ff4b4fa6b3b1648","value":"0xde0b6b3a7640000"}'`)
	cli.TestCommand(`rpc eth_estimateGas '{"data":"0xa9059cbb00000000000000000000000082a3a88bc9d6a70c4f3c66534566892eae0cad810000000000000000000000000000000000000000000000000000000000000001","from":"0x82a3a88bc9d6a70c4f3c66534566892eae0cad89","to":"0x0b7789b5f69678f4f2d237cd0e1c815e1cd39ccf","value":"0x0"}'`)
	cli.TestCommand("rpc eth_gasPrice")
	cli.TestCommand(`rpc eth_sendTransaction '{"from":"0xb60e8dd61c5d32be8058bb8eb970870f07233155","to":"0xd46e8dd67c5d32be8058bb8eb970870f07244567","gas":"0x76c0","gasPrice":"0x9184e72a000","value":"0x9184e72a","data":"0xd46e8dd67c5d32be8d46e8dd67c5d32be8058bb8eb970870f072445675058bb8eb970870f072445675"}'`)
}
