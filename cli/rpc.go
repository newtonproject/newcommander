package cli

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"
)

func (cli *CLI) buildRPCCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "rpc <jsonrpc_method> [jsonrpc_param1] [jsonrpc_param2]",
		Short:                 fmt.Sprintf("%s RPC method", cli.blockchain.String()),
		Args:                  cobra.MinimumNArgs(1),
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			method := args[0]

			var params []interface{}
			for _, arg := range args[1:] {
				var param interface{}
				if arg == "true" {
					param = true
				} else if arg == "false" {
					param = false
				} else if num, err := strconv.ParseUint(arg, 10, 0); err == nil {
					param = fmt.Sprintf(`0x%x`, num)
				} else if strings.HasPrefix(arg, "{") {
					if method == "eth_estimateGas" || method == "eth_call" {
						param = ethereum.CallMsg{}
						err := json.Unmarshal([]byte(arg), &param)
						if err != nil {
							fmt.Println("json unmarshal error: ", err)
							return
						}
					} else if method == "eth_sendTransaction" {
						param = types.Transaction{}
						err := json.Unmarshal([]byte(arg), &param)
						if err != nil {
							fmt.Println("json unmarshal error: ", err)
							return
						}
					} else {
						param = ""
					}
				} else {
					param = fmt.Sprintf(`%s`, arg)
				}

				params = append(params, param)
			}

			sendJSONPostAndShow(cli.rpcURL, method, params...)
		},
	}

	return cmd
}
