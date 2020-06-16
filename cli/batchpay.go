package cli

import (
	"bufio"
	"context"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/cobra"
)

func (cli *CLI) buildBatchPayCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "batchpay <batch.txt>",
		Aliases:               []string{"batch"},
		Short:                 "Batch pay base on file <batch.txt>",
		Args:                  cobra.MinimumNArgs(1),
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {

			batchFileName := args[0]
			file, err := os.Open(batchFileName)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer file.Close()

			// check from
			var address common.Address
			if cli.tran != nil {
				address = cli.tran.From
			}
			if cmd.Flags().Changed("from") {
				fromStr, err := cmd.Flags().GetString("from")
				if err != nil {
					fmt.Println(err)
					return
				}
				if !common.IsHexAddress(fromStr) {
					fmt.Println(errFromAddressIllegal)
					return
				}
				address = common.HexToAddress(fromStr)
			}
			if address == (common.Address{}) {
				fmt.Println("From address not set")
				return
			}
			wallet := keystore.NewKeyStore(cli.walletPath,
				keystore.StandardScryptN, keystore.StandardScryptP)
			if !wallet.HasAddress(address) {
				fmt.Println("From address not in wallet")
				return
			}

			client, err := ethclient.Dial(cli.rpcURL)
			if err != nil {
				fmt.Println(err)
				return
			}
			ctx := context.Background()

			chainID, err := client.NetworkID(ctx)
			if err != nil {
				fmt.Println(err)
				return
			}

			var data []byte
			if cmd.Flags().Changed("data") {
				dataStr, err := cmd.Flags().GetString("data")
				if err != nil {
					fmt.Println("Get data from flag error: ", err)
					return
				}
				data = []byte(dataStr)
			}

			gasPrice := big.NewInt(0)
			if cmd.Flags().Changed("price") {
				price, err := cmd.Flags().GetUint64("price")
				if err != nil {
					fmt.Println("price get error: ", err)
					return
				}
				gasPrice = big.NewInt(0).SetUint64(price)
			} else {
				gasPrice, err = client.SuggestGasPrice(ctx)
				if err != nil {
					fmt.Println("SuggestGasPrice error: ", err)
					return
				}
			}

			gasLimit := uint64(0)
			if cmd.Flags().Changed("gas") {
				gasLimit, err = cmd.Flags().GetUint64("gas")
				if err != nil {
					fmt.Println("gas limit error: ", err)
					return
				}
			}

			nonce := uint64(0)
			if cmd.Flags().Changed("nonce") {
				nonce, err = cmd.Flags().GetUint64("nonce")
				if err != nil {
					fmt.Println("nonce get error: ", err)
					return
				}
			} else {
				nonce, err = client.PendingNonceAt(ctx, address)
				if err != nil {
					fmt.Println("PendingNonceAt error: ", err)
					return
				}
			}

			totalAmount := big.NewInt(0)
			totalGas := big.NewInt(0)

			txList := make([]*types.Transaction, 0)
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				text := scanner.Text()
				l := strings.Split(text, ",")
				if len(l) != 2 {
					fmt.Println("parse error: ", text)
					return
				}

				var to common.Address
				if common.IsHexAddress(l[0]) {
					to = common.HexToAddress(l[0])
				} else {
					if cli.blockchain != NewChain {
						fmt.Println("Convert address error: ", l[0])
						return
					}
					to, err = newToAddress(chainID.Bytes(), l[0])
					if err != nil {
						fmt.Println("NewChain: address is invalid hex address or convert from NEW Address to hex error: ", l[0])
						return
					}
				}
				if to == (common.Address{}) {
					fmt.Println("Warning: to address is zero: ", l[0])
				}

				amount, err := getAmountWei(l[1], UnitETH)
				if err != nil {
					fmt.Println("amount error: ", err)
					return
				}

				if gasLimit == 0 {
					gasLimit, err = client.EstimateGas(ctx, ethereum.CallMsg{
						From:  address,
						To:    &to,
						Value: amount,
						Data:  data,
					})
					if err != nil {
						fmt.Println("EstimateGas error: ", err)
						return
					}
				}

				tx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, data)
				nonce++

				txList = append(txList, tx)

				// total
				totalAmount.Add(totalAmount, amount)
				totalGas.Add(totalGas, big.NewInt(0).Mul(gasPrice, big.NewInt(0).SetUint64(gasLimit)))
			}

			fmt.Println("Please confirm the transactions below:")
			for _, tx := range txList {
				// show info
				fmt.Printf("%s,%s\n", tx.To().String(), getWeiAmountTextByUnit(tx.Value(), UnitETH))
			}
			fmt.Println("Number of transactions:", len(txList))

			if totalAmount.Cmp(big.NewInt(0)) <= 0 {
				fmt.Println("Total pay amount is zero")
				return
			}

			balance, err := client.PendingBalanceAt(ctx, address)
			if err != nil {
				fmt.Println(err)
				return
			}

			if balance.Cmp(big.NewInt(0).Add(totalAmount, totalGas)) < 0 {
				fmt.Println("Error: Insufficient funds")
				return
			}

			fmt.Println("Total pay amount:", getWeiAmountTextUnitByUnit(totalAmount, UnitETH))
			fmt.Println("Total gas amount:", getWeiAmountTextUnitByUnit(totalGas, UnitETH))

			var walletPassword string
			if cli.tran != nil {
				walletPassword = cli.tran.Password
			}
			for trials := 0; trials <= 1; trials++ {
				err = wallet.Unlock(accounts.Account{Address: address}, walletPassword)
				if err == nil {
					break
				}
				if trials >= 1 {
					fmt.Printf("failed to unlock account %s (%v)\n", address.String(), err)
					return

				}
				prompt := fmt.Sprintf("Unlocking account %s", address.String())
				walletPassword, _ = getPassPhrase(prompt, false)
			}

			wait, _ := cmd.Flags().GetBool("wait")
			totalGasUsed := big.NewInt(0)
			for _, tx := range txList {
				signTx, err := wallet.SignTx(accounts.Account{Address: address}, tx, chainID)
				if err != nil {
					fmt.Println(err)
					return
				}

				err = client.SendTransaction(ctx, signTx)
				if err != nil {
					fmt.Println(err)
					return
				}

				fmt.Printf("Succeed broadcast pay %s to %s from %s with nonce %d, TxID %s.\n",
					getWeiAmountTextUnitByUnit(signTx.Value(), UnitETH),
					signTx.To().String(), address.String(),
					signTx.Nonce(), signTx.Hash().String())

				if wait {
					txr, err := bind.WaitMined(ctx, client, signTx)
					if err != nil {
						fmt.Println(err)
						return
					}
					if txr.Status == 1 {
						fmt.Printf("Succeed mined txID %s.\n", txr.TxHash.String())
					} else {
						fmt.Printf("Succeed mined txID %s but status failed.\n", txr.TxHash.String())
					}
					totalGasUsed.Add(totalGasUsed, big.NewInt(0).Mul(tx.GasPrice(), big.NewInt(0).SetUint64(txr.GasUsed)))
				} else {
					totalGasUsed.Add(totalGasUsed, big.NewInt(0).Mul(tx.GasPrice(), big.NewInt(0).SetUint64(tx.Gas())))
				}
			}

			if wait {
				fmt.Println("Total gas used amount:", getWeiAmountTextUnitByUnit(totalGasUsed, UnitETH))
			}
		},
	}

	cmd.Flags().String("from", "", "source account address")
	cmd.Flags().String("data", "", "custom data message (use quotes if there are spaces)")
	cmd.Flags().Uint64P("gas", "g", 21000, "the gas provided for the transaction execution")
	cmd.Flags().Uint64P("price", "p", 1, fmt.Sprintf("the gasPrice used for each paid gas (unit in %s)", UnitWEI))
	cmd.Flags().Uint64P("nonce", "n", 0, "the number of nonce")
	cmd.Flags().Bool("wait", false, "wait for transaction to mined")

	return cmd
}
