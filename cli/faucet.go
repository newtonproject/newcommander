package cli

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

func (cli *CLI) buildFaucetCmd() *cobra.Command {
	faucetCmd := &cobra.Command{
		Use:                   "faucet [address]",
		Short:                 "Get free money for address on NewChain TestNet",
		Args:                  cobra.MinimumNArgs(1),
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			var addressList []common.Address

			all, _ := cmd.Flags().GetBool("all")

			if all && len(args) <= 0 {
				if err := cli.openWallet(true); err != nil {
					fmt.Println(err)
					return
				}
				for _, account := range cli.wallet.Accounts() {
					addressList = append(addressList, account.Address)
				}
			} else {
				for _, addressStr := range args {
					if common.IsHexAddress(addressStr) {
						addressList = append(addressList, common.HexToAddress(addressStr))
					} else {
						fmt.Println("address illegal:", addressStr)
					}
				}
			}

			for _, address := range addressList {
				getFaucet(cli.rpcURL, address.String())
			}
		},
	}

	faucetCmd.Flags().Bool("all", false, "get faucet for all accounts under the wallet")

	return faucetCmd
}
