package cli

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

func (cli *CLI) buildBalanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   fmt.Sprintf("balance [-u %s] [-n pending] [-s] [address1] [address2]...", strings.Join(UnitList, "|")),
		Short:                 "Get balance of address",
		Args:                  cobra.MinimumNArgs(0),
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			nosum, _ := cmd.Flags().GetBool("nosum")

			cli.showBalance(cmd, args, !nosum)
		},
	}

	cmd.Flags().StringP("unit", "u", "", fmt.Sprintf("unit for balance. %s.", UnitString))
	cmd.Flags().BoolP("safe", "s", false, "enable safe mode to check balance (force use the block 3 block heights less than the latest)")
	cmd.Flags().StringP("number", "n", "latest", `the integer block number, or the string "latest", "earliest" or "pending"`)
	cmd.Flags().Bool("nosum", false,   `disable show sum info`)

	return cmd
}

func (cli *CLI) getBalance(address common.Address) (*big.Int, error) {
	if err := cli.BuildClient(); err != nil {
		return nil, err
	}
	return cli.client.BalanceAt(context.Background(), address, nil)
}

func (cli *CLI) getPendingBalance(address common.Address) (*big.Int, error) {
	if err := cli.BuildClient(); err != nil {
		return nil, err
	}
	return cli.client.PendingBalanceAt(context.Background(), address)
}

func (cli *CLI) showBalance(cmd *cobra.Command, args []string, showSum bool) {
	var err error

	unit, _ := cmd.Flags().GetString("unit")
	if unit != "" && !stringInSlice(unit, UnitList) {
		fmt.Printf("Unit(%s) for invalid. %s.\n", unit, UnitString)
		fmt.Fprint(os.Stderr, cmd.UsageString())
		return
	}

	safe, _ := cmd.Flags().GetBool("safe")

	pending := false
	latest := true
	number := big.NewInt(0)
	if !safe {
		if cmd.Flags().Changed("number") {
			numStr, err := cmd.Flags().GetString("number")
			if err != nil {
				fmt.Println("Error: arg number get error: ", err)
				return
			}
			switch numStr {
			case "pending":
				pending = true
				latest = false
			case "earliest":
				pending = false
				latest = false
				number = number.SetUint64(0)
			case "latest":
				latest = true
				pending = false
			default:
				pending = false
				latest = false
				var ok bool
				number, ok = number.SetString(numStr, 10)
				if !ok {
					fmt.Println("Error: arg number convert to big int error")
					return
				}
				if number.Cmp(big.NewInt(0)) < 0 {
					fmt.Println("Error: arg number is less than 0")
					return
				}
			}
		}
	}

	var addressList []common.Address

	if len(args) <= 0 {
		if err := cli.openWallet(true); err != nil {
			fmt.Println(err)
			return
		}

		for _, account := range cli.wallet.Accounts() {
			addressList = append(addressList, account.Address)
		}

	} else {
		for _, addressStr := range args {
			addressList = append(addressList, common.HexToAddress(addressStr))
		}
	}

	if err := cli.BuildClient(); err != nil {
		fmt.Println(err)
		return
	}
	ctx := context.Background()

	var blockNumber *big.Int
	if safe {
		latestHeader, err := cli.client.HeaderByNumber(ctx, nil)
		if err != nil {
			fmt.Println("HeaderByBlock error: ", err)
			return
		}
		if latestHeader == nil {
			fmt.Println("HeaderByBlock return nil")
			return
		}
		blockNumber = big.NewInt(0).Sub(latestHeader.Number, big.NewInt(3))
		fmt.Printf("Safe mode enable, check balance at block height %s while the latest is %s\n", blockNumber.String(), latestHeader.Number.String())
	} else if !latest && !pending {
		blockNumber = number
	}

	balanceSum := big.NewInt(0)
	for _, address := range addressList {
		var balance *big.Int
		if pending {
			balance, err = cli.client.PendingBalanceAt(ctx, address)
		} else {
			balance, err = cli.client.BalanceAt(ctx, address, blockNumber)
		}
		balanceSum.Add(balanceSum, balance)
		if err != nil {
			fmt.Println("Balance error:", err)
			return
		}
		fmt.Printf("Address[%s] Balance[%s]\n", address.Hex(), getWeiAmountTextUnitByUnit(balance, unit))
	}

	if showSum {
		fmt.Println("Number Of Accounts:", len(addressList))
		fmt.Println("Total Balance:", getWeiAmountTextUnitByUnit(balanceSum, unit))
	}

	return
}
