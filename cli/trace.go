package cli

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/newtonproject/newcommander/tracer"
	"github.com/spf13/cobra"
)

func (cli *CLI) buildTraceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "trace <txHash>",
		Short:                 "trace tx with hash and get internal txs",
		DisableFlagsInUseLine: true,
		Args:                  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			rpcclient, err := rpc.Dial(cli.rpcURL)
			if err != nil {
				fmt.Println(err)
				return
			}

			txHash := common.HexToHash(args[0])

			config := &tracer.TraceConfig{
				Tracer:  "",
				Timeout: nil,
				Reexec:  nil,
			}

			txs, err := tracer.TraceTransaction(rpcclient, context.Background(), txHash, config)
			if err != nil {
				fmt.Println("Trace error: ", err)
				return
			}
			for _, tx := range txs {
				txJson, err := tx.MarshalJSON()
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Println(string(txJson))
			}
		},
	}

	return cmd
}
